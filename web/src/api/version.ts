import { fetchAPI } from './client'

export interface VersionInfo {
  version: string
}

export async function getVersion(): Promise<VersionInfo> {
  return fetchAPI<VersionInfo>('/api/version')
}
