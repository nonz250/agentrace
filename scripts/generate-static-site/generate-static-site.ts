/**
 * 静的サイト生成スクリプト
 *
 * 起動中のサーバーとWebアプリから各ページをクロールし、
 * レンダリング済みHTMLをdocs/配下に保存する
 *
 * 前提条件:
 *   1. サーバーが起動していること (default: http://localhost:8080)
 *   2. Webアプリが起動していること (default: http://localhost:5173)
 *   3. 表示したいデータがDBに存在すること
 *
 * 使用方法:
 *   cd scripts/generate-static-site && npm install && npm run generate-static
 *
 * 環境変数:
 *   AGENTRACE_SERVER_URL - サーバーURL (default: http://localhost:8080)
 *   AGENTRACE_WEB_URL - WebアプリURL (default: http://localhost:5173)
 *   GITHUB_PAGES_BASE - GitHub Pages用のベースパス (default: /agentrace)
 */

import puppeteer, { Browser, Page } from "puppeteer";
import { mkdirSync, writeFileSync, cpSync, rmSync, existsSync, readFileSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";
import { execSync } from "child_process";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT_DIR = join(__dirname, "../..");
const DOCS_DIR = join(ROOT_DIR, "docs");
const WEB_DIR = join(ROOT_DIR, "web");
const API_DIR = join(DOCS_DIR, "api");

const SERVER_URL = process.env.AGENTRACE_SERVER_URL || "http://localhost:8080";
const WEB_URL = process.env.AGENTRACE_WEB_URL || "http://localhost:5173";
const GITHUB_PAGES_BASE = process.env.GITHUB_PAGES_BASE ?? "/agentrace";

interface Project {
  id: string;
  canonical_git_repository: string;
}

interface Session {
  id: string;
  project: Project | null;
}

interface Plan {
  id: string;
  project: Project | null;
}

async function fetchProjects(): Promise<Project[]> {
  const response = await fetch(`${SERVER_URL}/api/projects`);
  if (!response.ok) {
    console.warn("Failed to fetch projects, using empty list");
    return [];
  }
  const data = await response.json();
  return data.projects || [];
}

async function fetchSessions(): Promise<Session[]> {
  const response = await fetch(`${SERVER_URL}/api/sessions`);
  if (!response.ok) {
    console.warn("Failed to fetch sessions, using empty list");
    return [];
  }
  const data = await response.json();
  return data.sessions || [];
}

async function fetchPlans(): Promise<Plan[]> {
  const response = await fetch(`${SERVER_URL}/api/plans`);
  if (!response.ok) {
    console.warn("Failed to fetch plans, using empty list");
    return [];
  }
  const data = await response.json();
  return data.plans || [];
}

// APIレスポンスを静的JSONファイルとして保存
async function saveApiResponses(
  projects: Project[],
  sessions: Session[],
  plans: Plan[]
): Promise<void> {
  console.log("Saving API responses as static JSON...");

  // api/ ディレクトリを作成
  mkdirSync(API_DIR, { recursive: true });

  // /api/me - 未認証状態
  writeFileSync(
    join(API_DIR, "me.json"),
    JSON.stringify({ error: "unauthorized" })
  );

  // /api/projects
  const projectsResponse = await fetch(`${SERVER_URL}/api/projects`);
  if (projectsResponse.ok) {
    const data = await projectsResponse.json();
    writeFileSync(join(API_DIR, "projects.json"), JSON.stringify(data));
  }

  // 各プロジェクトの詳細
  const projectsDir = join(API_DIR, "projects");
  mkdirSync(projectsDir, { recursive: true });
  for (const project of projects) {
    const response = await fetch(`${SERVER_URL}/api/projects/${project.id}`);
    if (response.ok) {
      const data = await response.json();
      writeFileSync(join(projectsDir, `${project.id}.json`), JSON.stringify(data));
    }

    // プロジェクト別のsessions
    const sessionsResponse = await fetch(`${SERVER_URL}/api/sessions?project_id=${project.id}`);
    if (sessionsResponse.ok) {
      const data = await sessionsResponse.json();
      writeFileSync(join(projectsDir, `${project.id}-sessions.json`), JSON.stringify(data));
    }

    // プロジェクト別のplans
    const plansResponse = await fetch(`${SERVER_URL}/api/plans?project_id=${project.id}`);
    if (plansResponse.ok) {
      const data = await plansResponse.json();
      writeFileSync(join(projectsDir, `${project.id}-plans.json`), JSON.stringify(data));
    }
  }

  // /api/sessions
  const sessionsResponse = await fetch(`${SERVER_URL}/api/sessions`);
  if (sessionsResponse.ok) {
    const data = await sessionsResponse.json();
    writeFileSync(join(API_DIR, "sessions.json"), JSON.stringify(data));
  }

  // /api/plans
  const plansResponse = await fetch(`${SERVER_URL}/api/plans`);
  if (plansResponse.ok) {
    const data = await plansResponse.json();
    writeFileSync(join(API_DIR, "plans.json"), JSON.stringify(data));
  }

  // 各セッションの詳細
  const sessionsDir = join(API_DIR, "sessions");
  mkdirSync(sessionsDir, { recursive: true });
  for (const session of sessions) {
    const response = await fetch(`${SERVER_URL}/api/sessions/${session.id}`);
    if (response.ok) {
      const data = await response.json();
      writeFileSync(join(sessionsDir, `${session.id}.json`), JSON.stringify(data));
    }
  }

  // 各プランの詳細とイベント
  const plansDir = join(API_DIR, "plans");
  mkdirSync(plansDir, { recursive: true });
  for (const plan of plans) {
    const response = await fetch(`${SERVER_URL}/api/plans/${plan.id}`);
    if (response.ok) {
      const data = await response.json();
      writeFileSync(join(plansDir, `${plan.id}.json`), JSON.stringify(data));
    }

    // プランのイベント履歴
    const eventsResponse = await fetch(`${SERVER_URL}/api/plans/${plan.id}/events`);
    if (eventsResponse.ok) {
      const data = await eventsResponse.json();
      writeFileSync(join(plansDir, `${plan.id}-events.json`), JSON.stringify(data));
    }
  }

  console.log("  API responses saved");
}

interface CollectedData {
  urls: string[];
  projects: Project[];
  sessions: Session[];
  plans: Plan[];
}

async function collectUrlsAndData(): Promise<CollectedData> {
  const urls: string[] = [
    "/", // プロジェクト一覧
  ];

  // プロジェクト詳細とセッション・プラン一覧
  const projects = await fetchProjects();
  console.log(`  Found ${projects.length} projects`);
  for (const project of projects) {
    urls.push(`/projects/${project.id}`);
    urls.push(`/projects/${project.id}/sessions`);
    urls.push(`/projects/${project.id}/plans`);
  }

  // セッション詳細
  const sessions = await fetchSessions();
  console.log(`  Found ${sessions.length} sessions`);
  for (const session of sessions) {
    if (session.project) {
      urls.push(`/projects/${session.project.id}/sessions/${session.id}`);
    } else {
      console.log(`  Warning: Session ${session.id} has no project`);
    }
  }

  // プラン詳細
  const plans = await fetchPlans();
  console.log(`  Found ${plans.length} plans`);
  for (const plan of plans) {
    if (plan.project) {
      urls.push(`/projects/${plan.project.id}/plans/${plan.id}`);
    } else {
      console.log(`  Warning: Plan ${plan.id} has no project`);
    }
  }

  return {
    urls: [...new Set(urls)], // 重複を除去
    projects,
    sessions,
    plans,
  };
}

function urlToFilePath(url: string): string {
  if (url === "/") {
    return "index.html";
  }
  // /projects/abc -> projects/abc/index.html
  return join(url.slice(1), "index.html");
}

async function waitForPageLoad(page: Page): Promise<void> {
  // React のローディングが完了するまで待機
  // Spinner コンポーネントがなくなるまで待つ
  try {
    await page.waitForFunction(
      () => {
        const spinners = document.querySelectorAll('[class*="animate-spin"]');
        return spinners.length === 0;
      },
      { timeout: 10000 }
    );
  } catch {
    console.warn("  Warning: Timeout waiting for page load");
  }

  // 追加で少し待機（非同期レンダリング完了のため）
  await new Promise((resolve) => setTimeout(resolve, 500));
}

// fetchをオーバーライドして静的JSONを返すスクリプト
function createFetchOverrideScript(): string {
  return `<script>
(function() {
  const BASE = '${GITHUB_PAGES_BASE}';
  const originalFetch = window.fetch;

  // APIパスを静的JSONファイルパスに変換
  function apiToStaticPath(url) {
    const urlObj = new URL(url, location.origin);
    const path = urlObj.pathname;

    // /api/* パスのみ処理
    if (!path.startsWith('/api/')) return null;

    // /api/plans/:id/events -> /api/plans/:id-events.json
    const eventsMatch = path.match(/^\\/api\\/plans\\/([^/]+)\\/events$/);
    if (eventsMatch) {
      return BASE + '/api/plans/' + eventsMatch[1] + '-events.json';
    }

    // /api/sessions/:id -> /api/sessions/:id.json
    // /api/plans/:id -> /api/plans/:id.json
    // /api/projects/:id -> /api/projects/:id.json
    const detailMatch = path.match(/^\\/api\\/(sessions|plans|projects)\\/([^/]+)$/);
    if (detailMatch) {
      return BASE + '/api/' + detailMatch[1] + '/' + detailMatch[2] + '.json';
    }

    // /api/sessions?project_id=xxx -> /api/projects/xxx-sessions.json
    // /api/plans?project_id=xxx -> /api/projects/xxx-plans.json
    const projectFilterMatch = path.match(/^\\/api\\/(sessions|plans)$/);
    if (projectFilterMatch) {
      const projectId = urlObj.searchParams.get('project_id');
      if (projectId) {
        return BASE + '/api/projects/' + projectId + '-' + projectFilterMatch[1] + '.json';
      }
    }

    // /api/sessions -> /api/sessions.json
    // /api/plans -> /api/plans.json
    // /api/projects -> /api/projects.json
    // /api/me -> /api/me.json
    const listMatch = path.match(/^\\/api\\/(sessions|plans|projects|me)$/);
    if (listMatch) {
      return BASE + '/api/' + listMatch[1] + '.json';
    }

    return null;
  }

  window.fetch = function(input, init) {
    const url = typeof input === 'string' ? input : input.url;
    const staticPath = apiToStaticPath(url);

    if (staticPath) {
      // POST/PATCH/DELETE は読み取り専用エラーを返す
      if (init && init.method && init.method !== 'GET') {
        return Promise.resolve(new Response(
          JSON.stringify({ error: 'This is a read-only demo' }),
          { status: 403, headers: { 'Content-Type': 'application/json' } }
        ));
      }
      return originalFetch(staticPath);
    }

    return originalFetch.apply(this, arguments);
  };
})();
</script>`;
}

async function renderPage(
  browser: Browser,
  url: string
): Promise<string> {
  const page = await browser.newPage();

  try {
    await page.goto(`${WEB_URL}${url}`, {
      waitUntil: "networkidle0",
      timeout: 30000,
    });

    await waitForPageLoad(page);

    // HTMLを取得
    let html = await page.content();

    // ベースパスを調整（GitHub Pages用）
    // 相対パスをベースパス付きの絶対パスに変換
    if (GITHUB_PAGES_BASE) {
      html = html.replace(
        /(<script[^>]*src="|<link[^>]*href=")\/assets\//g,
        `$1${GITHUB_PAGES_BASE}/assets/`
      );

      // 内部リンクのhref属性にベースパスを追加
      html = html.replace(
        /href="\/([^"]*?)"/g,
        `href="${GITHUB_PAGES_BASE}/$1"`
      );
    }

    // Vite開発サーバー用のスクリプトを除去
    html = html.replace(/<script[^>]*src="[^"]*@vite[^"]*"[^>]*><\/script>/g, '');
    html = html.replace(/<script[^>]*src="[^"]*@react-refresh[^"]*"[^>]*><\/script>/g, '');
    html = html.replace(/<script[^>]*src="[^"]*\/src\/main\.tsx[^"]*"[^>]*><\/script>/g, '');
    html = html.replace(/<script[^>]*type="module"[^>]*>[^<]*@vite[^<]*<\/script>/g, '');
    html = html.replace(/<script[^>]*>[\s\S]*?RefreshRuntime[\s\S]*?<\/script>/g, '');
    // インラインのViteスクリプトも除去
    html = html.replace(/<script type="module">[\s\S]*?<\/script>/g, '');

    // 本番用のスクリプトとスタイルを取得して追加
    const { scripts, styles } = extractProductionAssets();

    // ベースパスを適用したスクリプトとスタイル
    let prodScripts = scripts;
    let prodStyles = styles;
    if (GITHUB_PAGES_BASE) {
      prodScripts = scripts.replace(/src="\/assets\//g, `src="${GITHUB_PAGES_BASE}/assets/`);
      prodStyles = styles.replace(/href="\/assets\//g, `href="${GITHUB_PAGES_BASE}/assets/`);
    }

    // 開発用スタイルを除去して本番用に置き換え
    html = html.replace(/<link[^>]*rel="stylesheet"[^>]*>/g, '');
    html = html.replace('</head>', prodStyles + '</head>');

    // 本番用スクリプトを追加
    html = html.replace('</body>', prodScripts + '</body>');

    // fetchオーバーライドスクリプトを注入（他のスクリプトより先に実行されるようにする）
    html = html.replace(
      '<head>',
      '<head>' + createFetchOverrideScript()
    );

    return html;
  } finally {
    await page.close();
  }
}

async function buildWebApp(): Promise<void> {
  console.log("Building web app...");
  const env = {
    ...process.env,
    VITE_BASE_PATH: GITHUB_PAGES_BASE || undefined,
  };
  execSync("npm run build", {
    cwd: WEB_DIR,
    stdio: "inherit",
    env,
  });
}

// ビルド済みHTMLからスクリプトとスタイルのタグを抽出
function extractProductionAssets(): { scripts: string; styles: string } {
  const distIndexPath = join(WEB_DIR, "dist", "index.html");
  const html = readFileSync(distIndexPath, "utf-8");

  // スクリプトタグを抽出
  const scriptMatches = html.match(/<script[^>]*src="[^"]*\.js"[^>]*><\/script>/g) || [];
  const scripts = scriptMatches.join("\n");

  // スタイルタグを抽出
  const styleMatches = html.match(/<link[^>]*rel="stylesheet"[^>]*>/g) || [];
  const styles = styleMatches.join("\n");

  return { scripts, styles };
}

async function copyAssets(): Promise<void> {
  const distDir = join(WEB_DIR, "dist");
  const assetsDir = join(distDir, "assets");

  if (existsSync(assetsDir)) {
    const targetAssetsDir = join(DOCS_DIR, "assets");
    mkdirSync(targetAssetsDir, { recursive: true });
    cpSync(assetsDir, targetAssetsDir, { recursive: true });
    console.log("Assets copied to docs/assets/");
  }

  // favicon等のルートファイルもコピー
  const rootFiles = ["favicon.ico", "vite.svg"];
  for (const file of rootFiles) {
    const src = join(distDir, file);
    if (existsSync(src)) {
      cpSync(src, join(DOCS_DIR, file));
    }
  }
}

function create404Html(): void {
  // GitHub Pages用の404.htmlを作成
  // SPAのクライアントサイドルーティングをサポートするためのリダイレクト
  const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Agentrace Demo</title>
  <script>
    // GitHub Pages SPA redirect hack
    // 404ページに来たらルートにリダイレクトして、元のパスをsessionStorageに保存
    sessionStorage.redirect = location.href;
    location.replace(location.origin + '${GITHUB_PAGES_BASE}/');
  </script>
</head>
<body>
  <p>Redirecting...</p>
</body>
</html>`;

  writeFileSync(join(DOCS_DIR, "404.html"), html);
  console.log("Created 404.html for SPA routing");
}

function createNojekyll(): void {
  // GitHub Pagesで_から始まるファイルを無視しないようにする
  writeFileSync(join(DOCS_DIR, ".nojekyll"), "");
  console.log("Created .nojekyll");
}

async function main() {
  console.log("=== Agentrace Static Site Generator ===\n");
  console.log(`Server URL: ${SERVER_URL}`);
  console.log(`Web URL: ${WEB_URL}`);
  console.log(`GitHub Pages Base: ${GITHUB_PAGES_BASE}`);
  console.log(`Output: ${DOCS_DIR}\n`);

  // docs/をクリアして再作成
  if (existsSync(DOCS_DIR)) {
    rmSync(DOCS_DIR, { recursive: true });
  }
  mkdirSync(DOCS_DIR, { recursive: true });

  // Webアプリをビルド
  await buildWebApp();

  // アセットをコピー
  await copyAssets();

  // ブラウザを起動
  console.log("\nLaunching browser...");
  const browser = await puppeteer.launch({
    headless: true,
    args: ["--no-sandbox", "--disable-setuid-sandbox"],
  });

  try {
    // 利用可能なURLとデータを収集
    console.log("Collecting URLs from API...");
    const { urls, projects, sessions, plans } = await collectUrlsAndData();
    console.log(`Found ${urls.length} pages to render\n`);

    // APIレスポンスを静的JSONとして保存
    await saveApiResponses(projects, sessions, plans);

    // 各ページをレンダリング
    console.log("Rendering pages...");
    for (const url of urls) {
      const filePath = urlToFilePath(url);
      const fullPath = join(DOCS_DIR, filePath);

      console.log(`  ${url} -> ${filePath}`);

      // ディレクトリを作成
      mkdirSync(dirname(fullPath), { recursive: true });

      // ページをレンダリングして保存
      const html = await renderPage(browser, url);
      writeFileSync(fullPath, html);
    }

    // 404.htmlを作成
    create404Html();

    // .nojekyllを作成
    createNojekyll();

    console.log("\n=== Static site generation complete! ===");
    console.log(`Output: ${DOCS_DIR}`);
    console.log(`\nTo preview locally:`);
    console.log(`  npx serve docs`);
    console.log(`\nTo deploy to GitHub Pages:`);
    console.log(`  1. Push the docs/ folder to your repository`);
    console.log(`  2. Enable GitHub Pages in repository settings`);
    console.log(`  3. Set source to "Deploy from a branch" and select the docs folder`);
  } finally {
    await browser.close();
  }
}

main().catch((err) => {
  console.error("Error:", err);
  process.exit(1);
});
