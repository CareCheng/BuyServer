import nextConfig from 'eslint-config-next/core-web-vitals'

const config = [
  { ignores: ['**/.next/**', '**/out/**', '**/node_modules/**'] },
  ...nextConfig,
  {
    rules: {
      '@next/next/no-img-element': 'off',
      '@next/next/no-html-link-for-pages': 'off',
      'react-hooks/immutability': 'off',
      'react-hooks/set-state-in-effect': 'off',
      'react-hooks/exhaustive-deps': 'off',
      'react/no-unescaped-entities': 'off',
    },
  },
]

export default config
