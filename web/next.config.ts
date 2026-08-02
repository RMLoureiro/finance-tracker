import type { NextConfig } from "next";

// Em produção a app é exportada estaticamente e servida pelo binário Go;
// em desenvolvimento o Next faz proxy de /api para a API Go em :8080.
const isExport = process.env.BUILD_EXPORT === "1";

const nextConfig: NextConfig = {
  ...(isExport ? { output: "export" as const } : {}),
  async rewrites() {
    if (isExport) return [];
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/api/:path*",
      },
    ];
  },
};

export default nextConfig;
