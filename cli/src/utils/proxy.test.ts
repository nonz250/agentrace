import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ProxyAgent } from 'undici';

// モックを設定
vi.mock('../config/manager.js', () => ({
  loadConfig: vi.fn(),
}));

import { getProxyUrl, createDispatcher } from './proxy.js';
import { loadConfig } from '../config/manager.js';

const mockedLoadConfig = vi.mocked(loadConfig);

describe('proxy', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    // 環境変数をリセット
    process.env = { ...originalEnv };
    delete process.env.HTTPS_PROXY;
    delete process.env.https_proxy;
    delete process.env.HTTP_PROXY;
    delete process.env.http_proxy;

    // モックをリセット
    vi.clearAllMocks();
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  describe('getProxyUrl', () => {
    it('設定ファイルにproxy_urlがある場合、それを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
        proxy_url: 'http://proxy.example.com:8080',
      });

      expect(getProxyUrl()).toBe('http://proxy.example.com:8080');
    });

    it('設定ファイルにproxy_urlがなく、HTTPS_PROXYがある場合、それを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
      });
      process.env.HTTPS_PROXY = 'http://https-proxy.example.com:8080';

      expect(getProxyUrl()).toBe('http://https-proxy.example.com:8080');
    });

    it('設定ファイルにproxy_urlがなく、https_proxy（小文字）がある場合、それを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
      });
      process.env.https_proxy = 'http://https-proxy-lower.example.com:8080';

      expect(getProxyUrl()).toBe('http://https-proxy-lower.example.com:8080');
    });

    it('設定ファイルにproxy_urlがなく、HTTP_PROXYがある場合、それを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
      });
      process.env.HTTP_PROXY = 'http://http-proxy.example.com:8080';

      expect(getProxyUrl()).toBe('http://http-proxy.example.com:8080');
    });

    it('設定ファイルにproxy_urlがなく、http_proxy（小文字）がある場合、それを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
      });
      process.env.http_proxy = 'http://http-proxy-lower.example.com:8080';

      expect(getProxyUrl()).toBe('http://http-proxy-lower.example.com:8080');
    });

    it('設定ファイルも環境変数もない場合、undefinedを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
      });

      expect(getProxyUrl()).toBeUndefined();
    });

    it('設定ファイルがnullの場合、環境変数にフォールバックする', () => {
      mockedLoadConfig.mockReturnValue(null);
      process.env.HTTPS_PROXY = 'http://env-proxy.example.com:8080';

      expect(getProxyUrl()).toBe('http://env-proxy.example.com:8080');
    });

    it('設定ファイルが環境変数より優先される', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
        proxy_url: 'http://config-proxy.example.com:8080',
      });
      process.env.HTTPS_PROXY = 'http://env-proxy.example.com:8080';

      expect(getProxyUrl()).toBe('http://config-proxy.example.com:8080');
    });

    it('HTTPS_PROXYがHTTP_PROXYより優先される', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
      });
      process.env.HTTPS_PROXY = 'http://https-proxy.example.com:8080';
      process.env.HTTP_PROXY = 'http://http-proxy.example.com:8080';

      expect(getProxyUrl()).toBe('http://https-proxy.example.com:8080');
    });
  });

  describe('createDispatcher', () => {
    it('プロキシURLがある場合、ProxyAgentを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
        proxy_url: 'http://proxy.example.com:8080',
      });

      const dispatcher = createDispatcher();
      expect(dispatcher).toBeInstanceOf(ProxyAgent);
    });

    it('プロキシURLがない場合、undefinedを返す', () => {
      mockedLoadConfig.mockReturnValue({
        server_url: 'http://localhost:8080',
        api_key: 'test-key',
      });

      const dispatcher = createDispatcher();
      expect(dispatcher).toBeUndefined();
    });
  });
});
