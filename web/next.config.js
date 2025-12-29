/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  trailingSlash: true,
  assetPrefix: '/static',
  images: {
    unoptimized: true,
  },
}

module.exports = nextConfig
