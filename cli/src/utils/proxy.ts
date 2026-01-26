import { ProxyAgent } from 'undici';
import { loadConfigWithFallback } from '../config/manager.js';

/**
 * プロキシURLを取得する
 * 優先順位: 設定ファイル > 環境変数（HTTPS_PROXY > HTTP_PROXY）
 */
export function getProxyUrl(projectDir?: string): string | undefined {
  const config = loadConfigWithFallback(projectDir);

  // 設定ファイル優先
  if (config?.proxy_url) {
    return config.proxy_url;
  }

  // 環境変数フォールバック（大文字・小文字両方サポート）
  return process.env.HTTPS_PROXY
    || process.env.https_proxy
    || process.env.HTTP_PROXY
    || process.env.http_proxy;
}

/**
 * undici の dispatcher を生成する
 * プロキシ設定がない場合は undefined を返す（デフォルト動作）
 */
export function createDispatcher(projectDir?: string): ProxyAgent | undefined {
  const proxyUrl = getProxyUrl(projectDir);

  if (!proxyUrl) {
    return undefined;
  }

  return new ProxyAgent(proxyUrl);
}
