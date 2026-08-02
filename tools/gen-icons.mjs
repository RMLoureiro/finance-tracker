// Gera os PNG do manifest (192/512/apple-touch 180) a partir do desenho do icon.svg.
import { chromium } from "playwright";

const OUT = "../web/public";

const svg = (rx) => `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" width="100%" height="100%">
  <rect width="512" height="512" rx="${rx}" fill="#0f9488"/>
  <rect x="96" y="152" width="320" height="224" rx="40" fill="none" stroke="#ffffff" stroke-width="30"/>
  <path d="M96 216h320" stroke="#ffffff" stroke-width="30"/>
  <circle cx="344" cy="296" r="22" fill="#ffffff"/>
</svg>`;

// rx em unidades do viewBox (0–512); iOS aplica a sua própria máscara,
// por isso o apple-touch-icon vai com cantos retos e fundo opaco
const jobs = [
  { name: "icon-192.png", size: 192, rx: 112, transparent: true },
  { name: "icon-512.png", size: 512, rx: 112, transparent: true },
  { name: "apple-touch-icon.png", size: 180, rx: 0, transparent: false },
];

const browser = await chromium.launch();
for (const { name, size, rx, transparent } of jobs) {
  const page = await browser.newPage({ viewport: { width: size, height: size } });
  await page.setContent(
    `<!doctype html><body style="margin:0;background:transparent">${svg(rx)}</body>`
  );
  await page.screenshot({ path: `${OUT}/${name}`, omitBackground: transparent });
  await page.close();
  console.log("ok", name);
}
await browser.close();
