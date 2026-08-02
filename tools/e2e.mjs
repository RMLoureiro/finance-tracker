// Percurso end-to-end da app «Bolso» contra o build de produção servido pelo Go.
import { chromium } from "playwright";

const BASE = process.env.BASE_URL ?? "http://localhost:8091";
const SHOTS = process.env.SHOTS_DIR ?? "./shots";

const errors = [];
const browser = await chromium.launch();
const ctx = await browser.newContext({ viewport: { width: 390, height: 844 } });
const page = await ctx.newPage();
page.on("pageerror", (e) => errors.push(`pageerror: ${e.message}`));
page.on("console", (m) => {
  if (m.type() === "error") errors.push(`console: ${m.text()}`);
});

async function shot(name) {
  await page.screenshot({ path: `${SHOTS}/${name}.png`, fullPage: false });
}

function fail(msg) {
  console.error("FALHOU:", msg);
  process.exitCode = 1;
}

// 1. primeiro arranque → redireciona para /login em modo setup
await page.goto(BASE + "/");
await page.waitForURL("**/login", { timeout: 15000 });
await page.waitForLoadState("networkidle");
if (!(await page.getByText("Primeira utilização", { exact: false }).isVisible()))
  fail("modo de setup não apareceu");
await shot("01-setup");

await page.fill("#lg-user", "loureiro");
await page.fill("#lg-pass", "palavra-passe-de-teste");
await page.fill("#lg-confirm", "palavra-passe-de-teste");
await page.getByRole("button", { name: "Criar acesso" }).click();
await page.waitForURL(BASE + "/", { timeout: 15000 });
await page.waitForLoadState("networkidle");

// 2. definir orçamentos (Água 30 €, Supermercado 320 €)
await page.goto(BASE + "/orcamentos");
await page.waitForLoadState("networkidle");
for (const [cat, valor] of [["Água", "30"], ["Supermercado", "320"]]) {
  await page.getByRole("button", { name: new RegExp(cat) }).first().click();
  await page.fill("#bg-monthly", valor);
  await page.getByRole("button", { name: "Guardar" }).click();
  await page.waitForSelector(".sheet", { state: "detached" });
}

// 3. despesa fixa mensal
await page.getByRole("button", { name: "Adicionar despesa fixa" }).click();
await page.fill("#rc-desc", "Fatura da água");
await page.selectOption("#rc-cat", { label: "💧 Água" });
await page.fill("#rc-amount", "28");
await page.fill("#rc-day", "12");
await page.getByRole("button", { name: "Guardar" }).click();
await page.waitForSelector(".sheet", { state: "detached" });
await page.waitForLoadState("networkidle");
await shot("02-orcamentos");

// 4. início: prevista visível + registar despesa manual
await page.goto(BASE + "/");
await page.waitForLoadState("networkidle");
if (!(await page.getByText("Fatura da água").first().isVisible()))
  fail("despesa prevista não aparece no Início");

await page.getByRole("button", { name: "+ Despesa" }).click();
await page.fill("#exp-amount", "12,50");
await page.selectOption("#exp-cat", { label: "🛒 Supermercado" });
await page.fill("#exp-desc", "Continente");
await page.getByRole("button", { name: "Guardar" }).click();
await page.waitForSelector(".sheet", { state: "detached" });
await page.waitForLoadState("networkidle");
if (!(await page.getByText("Continente").first().isVisible()))
  fail("despesa registada não aparece nos Recentes");

// 5. confirmar a prevista com o valor real
await page.getByRole("button", { name: "Confirmar" }).first().click();
await page.fill("#exp-amount", "27,80");
await page.getByRole("button", { name: "Confirmar" }).last().click();
await page.waitForSelector(".sheet", { state: "detached" });
await page.waitForLoadState("networkidle");
await shot("03-inicio");

// o disponível deve refletir 350 − 12,50 − 27,80 = 309,70 €
const header = await page.locator(".head").innerText();
if (!header.includes("309,70")) fail(`disponível errado no cabeçalho: ${header.replace(/\n/g, " | ")}`);

// 6. movimentos
await page.goto(BASE + "/movimentos");
await page.waitForLoadState("networkidle");
for (const txt of ["Continente", "Fatura da água"]) {
  if (!(await page.getByText(txt).first().isVisible())) fail(`«${txt}» não aparece nos movimentos`);
}
await shot("04-movimentos");

// 7. relatórios
await page.goto(BASE + "/relatorios");
await page.waitForLoadState("networkidle");
if ((await page.locator("svg[role=img]").count()) === 0) fail("gráfico dos relatórios não renderizou");
await shot("05-relatorios");

// 8. desktop (com a sessão do contexto mobile)
const desktop = await browser.newContext({
  viewport: { width: 1280, height: 900 },
  storageState: await ctx.storageState(),
});
const dpage = await desktop.newPage();
await dpage.goto(BASE + "/");
await dpage.waitForLoadState("networkidle");
await dpage.screenshot({ path: `${SHOTS}/06-desktop.png` });
await desktop.close();

// 9. sessão: logout → /login
await page.goto(BASE + "/orcamentos");
await page.waitForLoadState("networkidle");
await page.getByRole("button", { name: "Terminar sessão" }).click();
await page.waitForURL("**/login", { timeout: 15000 });

// os 401 antes do login são esperados: a app pede /api, recebe 401 e
// redireciona para /login
const relevant = errors.filter((e) => !e.includes("favicon") && !e.includes("401"));
if (relevant.length) fail("erros de consola:\n" + relevant.join("\n"));

await browser.close();
console.log(process.exitCode ? "E2E COM FALHAS" : "E2E OK");
