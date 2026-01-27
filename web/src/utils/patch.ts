import DiffMatchPatch from 'diff-match-patch'
import type { PlanDocumentEvent } from '@/types/plan-document'

/**
 * パッチがdiff-match-patchのネイティブ形式かどうかを判定
 * ネイティブ形式は `@@ -start,len +start,len @@` で始まる
 */
function isDmpNativeFormat(patch: string): boolean {
  return patch.startsWith('@@ ')
}

/**
 * サーバーの初期パッチ形式（全行+プレフィックス）からテキストを復元
 * 例: "+line1\n+line2\n+line3" -> "line1\nline2\nline3"
 */
function applyInitialPatch(patch: string): string {
  const lines = patch.split('\n')
  const contentLines: string[] = []

  for (const line of lines) {
    if (line.startsWith('+')) {
      contentLines.push(line.slice(1))
    }
  }

  return contentLines.join('\n')
}

/**
 * diff-match-patchのネイティブ形式パッチを適用
 */
function applyDmpPatch(dmp: DiffMatchPatch, content: string, patch: string): string {
  const patches = dmp.patch_fromText(patch)
  const [newContent] = dmp.patch_apply(patches, content)
  return newContent
}

/**
 * 初期パッチから指定イベントまでの全パッチを適用してテキストを復元
 * @param events body_changeイベントのみ（時系列順）
 * @param targetEventIndex 復元したいバージョンのインデックス（0-based）
 * @returns 復元されたテキスト
 */
export function reconstructContent(
  events: PlanDocumentEvent[],
  targetEventIndex: number
): string {
  const dmp = new DiffMatchPatch()
  let content = ''

  for (let i = 0; i <= targetEventIndex; i++) {
    const event = events[i]
    if (!event.patch) continue

    try {
      if (i === 0 && !isDmpNativeFormat(event.patch)) {
        // 初期パッチ（サーバーの独自形式）
        content = applyInitialPatch(event.patch)
      } else if (isDmpNativeFormat(event.patch)) {
        // diff-match-patchのネイティブ形式
        content = applyDmpPatch(dmp, content, event.patch)
      } else {
        // フォールバック: 初期パッチ形式として処理
        content = applyInitialPatch(event.patch)
      }
    } catch (e) {
      console.error(`Failed to apply patch at index ${i}:`, e)
    }
  }

  return content
}
