import type { Metadata, Viewport } from "next";
import "./globals.css";
import TabBar from "@/components/TabBar";
import SWRegister from "@/components/SWRegister";

export const metadata: Metadata = {
  title: "Bolso",
  description: "Contas da casa: orçamentos, despesas e relatórios.",
  manifest: "/manifest.webmanifest",
  appleWebApp: {
    capable: true,
    title: "Bolso",
    statusBarStyle: "black-translucent",
  },
  icons: {
    icon: "/icon.svg",
    apple: "/apple-touch-icon.png",
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#1d2426" },
    { media: "(prefers-color-scheme: dark)", color: "#0c1110" },
  ],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="pt-PT">
      <body>
        <div className="shell">
          {children}
          <TabBar />
        </div>
        <SWRegister />
      </body>
    </html>
  );
}
