package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// GenerateECCKeyPair 生成ECC密钥对 (P-256, 256位)
func GenerateECCKeyPair() (privateKeyPEM, publicKeyPEM string, err error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	// 编码私钥
	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}
	privBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	}
	privateKeyPEM = string(pem.EncodeToMemory(privBlock))

	// 编码公钥
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	publicKeyPEM = string(pem.EncodeToMemory(pubBlock))

	return privateKeyPEM, publicKeyPEM, nil
}

// ParseECCPrivateKey 解析ECC私钥
func ParseECCPrivateKey(pemData string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

// ParseECCPublicKey 解析ECC公钥
func ParseECCPublicKey(pemData string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}
	return ecdsaPub, nil
}

// ECCSign 使用ECC私钥签名
func ECCSign(privateKeyPEM string, data []byte) (string, error) {
	privateKey, err := ParseECCPrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", err
	}

	// 将r和s编码为固定长度的字节（与服务端保持一致）
	keySize := (privateKey.Curve.Params().BitSize + 7) / 8
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// 填充到固定长度
	signature := make([]byte, 2*keySize)
	copy(signature[keySize-len(rBytes):keySize], rBytes)
	copy(signature[2*keySize-len(sBytes):], sBytes)

	return base64.StdEncoding.EncodeToString(signature), nil
}

// ECCVerify 使用ECC公钥验证签名
func ECCVerify(publicKeyPEM string, data []byte, signatureB64 string) (bool, error) {
	publicKey, err := ParseECCPublicKey(publicKeyPEM)
	if err != nil {
		return false, err
	}

	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, err
	}

	// 分割r和s
	keySize := (publicKey.Curve.Params().BitSize + 7) / 8
	if len(signature) != 2*keySize {
		return false, errors.New("invalid signature length")
	}

	r := new(big.Int).SetBytes(signature[:keySize])
	s := new(big.Int).SetBytes(signature[keySize:])

	hash := sha256.Sum256(data)
	return ecdsa.Verify(publicKey, hash[:], r, s), nil
}

// ECCEncrypt 使用ECIES加密 (ECDH + AES-GCM)
func ECCEncrypt(publicKeyPEM string, plaintext []byte) (string, error) {
	publicKey, err := ParseECCPublicKey(publicKeyPEM)
	if err != nil {
		return "", err
	}

	// 生成临时密钥对
	ephemeralPrivate, err := ecdsa.GenerateKey(publicKey.Curve, rand.Reader)
	if err != nil {
		return "", err
	}

	// ECDH共享密钥
	x, _ := publicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, ephemeralPrivate.D.Bytes())
	sharedKey := sha256.Sum256(x.Bytes())

	// 使用AES-GCM加密
	block, err := aes.NewCipher(sharedKey[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// 编码临时公钥
	ephemeralPubBytes, _ := x509.MarshalPKIXPublicKey(&ephemeralPrivate.PublicKey)

	// 组合: 临时公钥长度(2字节) + 临时公钥 + nonce长度(1字节) + nonce + 密文
	result := make([]byte, 2+len(ephemeralPubBytes)+1+len(nonce)+len(ciphertext))
	result[0] = byte(len(ephemeralPubBytes) >> 8)
	result[1] = byte(len(ephemeralPubBytes))
	copy(result[2:], ephemeralPubBytes)
	result[2+len(ephemeralPubBytes)] = byte(len(nonce))
	copy(result[3+len(ephemeralPubBytes):], nonce)
	copy(result[3+len(ephemeralPubBytes)+len(nonce):], ciphertext)

	return base64.StdEncoding.EncodeToString(result), nil
}

// ECCDecrypt 使用ECIES解密
// 支持两种格式：
// 1. XOR流加密格式（与Server端兼容）：临时公钥长度(2字节) + 临时公钥 + 密文
// 2. AES-GCM格式：临时公钥长度(2字节) + 临时公钥 + nonce长度(1字节) + nonce + 密文
func ECCDecrypt(privateKeyPEM string, encryptedB64 string) ([]byte, error) {
	privateKey, err := ParseECCPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	encrypted, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return nil, err
	}

	if len(encrypted) < 3 {
		return nil, errors.New("加密数据无效")
	}

	// 解析临时公钥长度
	pubKeyLen := int(encrypted[0])<<8 | int(encrypted[1])
	if len(encrypted) < 2+pubKeyLen {
		return nil, errors.New("加密数据无效")
	}

	// 解析临时公钥
	ephemeralPubBytes := encrypted[2 : 2+pubKeyLen]
	ephemeralPub, err := x509.ParsePKIXPublicKey(ephemeralPubBytes)
	if err != nil {
		return nil, err
	}
	ephemeralECDSA, ok := ephemeralPub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("临时公钥无效")
	}

	// ECDH共享密钥
	x, _ := privateKey.Curve.ScalarMult(ephemeralECDSA.X, ephemeralECDSA.Y, privateKey.D.Bytes())
	sharedKey := sha256.Sum256(x.Bytes())

	// 获取密文部分
	ciphertext := encrypted[2+pubKeyLen:]

	// 判断加密格式：检查是否有足够的数据包含 nonce
	// AES-GCM 格式需要至少 1(nonce长度) + 12(nonce) + 16(tag) = 29 字节
	// 如果密文长度小于 29 字节，或者第一个字节不是有效的 nonce 长度（通常是 12），则使用 XOR 格式
	if len(ciphertext) > 0 {
		possibleNonceLen := int(ciphertext[0])
		// AES-GCM 的 nonce 通常是 12 字节
		if possibleNonceLen == 12 && len(ciphertext) >= 1+possibleNonceLen+16 {
			// 尝试 AES-GCM 解密
			nonce := ciphertext[1 : 1+possibleNonceLen]
			aesCiphertext := ciphertext[1+possibleNonceLen:]

			block, err := aes.NewCipher(sharedKey[:])
			if err == nil {
				gcm, err := cipher.NewGCM(block)
				if err == nil {
					plaintext, err := gcm.Open(nil, nonce, aesCiphertext, nil)
					if err == nil {
						return plaintext, nil
					}
				}
			}
			// AES-GCM 解密失败，回退到 XOR 格式
		}
	}

	// 使用 XOR 流解密（与 Server 端兼容）
	plaintext := make([]byte, len(ciphertext))
	for i := range ciphertext {
		plaintext[i] = ciphertext[i] ^ sharedKey[i%32]
	}

	return plaintext, nil
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// GenerateNonce 生成防重放随机数（16字节，返回32位十六进制字符串）
func GenerateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// 配置加密密钥（动态密钥，从配置数据库加载）
var configEncryptionKey []byte

// 默认密钥（仅在密钥未初始化时使用）
var defaultEncryptionKey = []byte("UserFrontendKey!") // 16字节默认密钥

// SetConfigEncryptionKey 设置配置加密密钥
func SetConfigEncryptionKey(key []byte) {
	configEncryptionKey = key
}

// GetConfigEncryptionKey 获取当前配置加密密钥
func GetConfigEncryptionKey() []byte {
	if len(configEncryptionKey) == 0 {
		return defaultEncryptionKey
	}
	return configEncryptionKey
}

// GenerateAESKey 生成指定长度的AES密钥
// keyLength: 128, 192, 或 256 位
func GenerateAESKey(keyLength int) (string, error) {
	var keyBytes int
	switch keyLength {
	case 128:
		keyBytes = 16
	case 192:
		keyBytes = 24
	case 256:
		keyBytes = 32
	default:
		keyBytes = 32 // 默认256位
	}

	key := make([]byte, keyBytes)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

// AESEncrypt 使用AES-GCM加密字符串（用于配置加密）
func AESEncrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := GetConfigEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt 使用AES-GCM解密字符串（用于配置解密）
func AESDecrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	key := GetConfigEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// AESEncryptWithKey 使用指定密钥加密
func AESEncryptWithKey(plaintext string, keyBase64 string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecryptWithKey 使用指定密钥解密
func AESDecryptWithKey(encrypted string, keyBase64 string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IsEncrypted 检查字符串是否已加密（通过尝试base64解码和长度判断）
func IsEncrypted(s string) bool {
	if s == "" {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}
	// AES-GCM加密后的数据至少包含nonce(12字节)+tag(16字节)
	return len(decoded) >= 28
}
