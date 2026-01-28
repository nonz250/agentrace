import DiffMatchPatch from 'diff-match-patch'

export interface DiffLine {
  type: 'unchanged' | 'deleted' | 'added'
  content: string
  beforeLineNo?: number
  afterLineNo?: number
}

export interface DiffHunk {
  beforeStartLine: number
  afterStartLine: number
  lines: DiffLine[]
}

/**
 * 行単位のdiffを計算し、ハンクにまとめる
 * @param before 変更前のテキスト
 * @param after 変更後のテキスト
 * @param contextLines 変更箇所の前後に表示するコンテキスト行数
 * @param mergeGap この行数以内の変更は1つのハンクにマージ
 */
export function computeLineDiff(
  before: string,
  after: string,
  contextLines: number = 5,
  mergeGap: number = 5
): DiffHunk[] {
  const dmp = new DiffMatchPatch()

  // 行単位でdiffを計算するため、行を文字に変換
  const lineToCharResult = dmp.diff_linesToChars_(before, after)
  const chars1 = lineToCharResult.chars1
  const chars2 = lineToCharResult.chars2
  const lineArray = lineToCharResult.lineArray

  // 文字レベルでdiff計算（実際は行レベル）
  const diffs = dmp.diff_main(chars1, chars2, false)

  // 文字を行に戻す
  dmp.diff_charsToLines_(diffs, lineArray)

  // DiffLine配列に変換
  const allLines: DiffLine[] = []
  let beforeLineNo = 1
  let afterLineNo = 1

  for (const [op, text] of diffs) {
    // 末尾の改行を除去して行に分割
    const lines = text.replace(/\n$/, '').split('\n')

    for (const line of lines) {
      if (op === 0) {
        // 変更なし
        allLines.push({
          type: 'unchanged',
          content: line,
          beforeLineNo: beforeLineNo++,
          afterLineNo: afterLineNo++,
        })
      } else if (op === -1) {
        // 削除
        allLines.push({
          type: 'deleted',
          content: line,
          beforeLineNo: beforeLineNo++,
        })
      } else if (op === 1) {
        // 追加
        allLines.push({
          type: 'added',
          content: line,
          afterLineNo: afterLineNo++,
        })
      }
    }
  }

  // 変更箇所のインデックスを特定
  const changeIndices: number[] = []
  allLines.forEach((line, index) => {
    if (line.type !== 'unchanged') {
      changeIndices.push(index)
    }
  })

  if (changeIndices.length === 0) {
    return []
  }

  // 変更箇所をハンクにグループ化（コンテキスト付き）
  const hunks: DiffHunk[] = []
  let currentHunkStart = Math.max(0, changeIndices[0] - contextLines)
  let currentHunkEnd = Math.min(allLines.length - 1, changeIndices[0] + contextLines)

  for (let i = 1; i < changeIndices.length; i++) {
    const changeStart = changeIndices[i] - contextLines
    const changeEnd = changeIndices[i] + contextLines

    // 前のハンクとマージするかどうか
    if (changeStart <= currentHunkEnd + mergeGap) {
      // マージ
      currentHunkEnd = Math.min(allLines.length - 1, changeEnd)
    } else {
      // 新しいハンク
      hunks.push(createHunk(allLines, currentHunkStart, currentHunkEnd))
      currentHunkStart = Math.max(0, changeStart)
      currentHunkEnd = Math.min(allLines.length - 1, changeEnd)
    }
  }

  // 最後のハンクを追加
  hunks.push(createHunk(allLines, currentHunkStart, currentHunkEnd))

  return hunks
}

function createHunk(allLines: DiffLine[], start: number, end: number): DiffHunk {
  const lines = allLines.slice(start, end + 1)

  // beforeStartLine と afterStartLine を計算
  let beforeStartLine = 1
  let afterStartLine = 1

  for (let i = 0; i < start; i++) {
    const line = allLines[i]
    if (line.type === 'unchanged' || line.type === 'deleted') {
      beforeStartLine++
    }
    if (line.type === 'unchanged' || line.type === 'added') {
      afterStartLine++
    }
  }

  return {
    beforeStartLine,
    afterStartLine,
    lines,
  }
}

/**
 * ハンク間の省略行数を計算
 */
export function getSkippedLines(
  hunks: DiffHunk[],
  hunkIndex: number
): { before: number; after: number } | null {
  if (hunkIndex === 0) {
    // 最初のハンクの前
    const hunk = hunks[0]
    if (hunk.beforeStartLine > 1 || hunk.afterStartLine > 1) {
      return {
        before: hunk.beforeStartLine - 1,
        after: hunk.afterStartLine - 1,
      }
    }
    return null
  }

  const prevHunk = hunks[hunkIndex - 1]
  const currentHunk = hunks[hunkIndex]

  // 前のハンクの最後の行番号を計算
  let prevBeforeEnd = prevHunk.beforeStartLine
  let prevAfterEnd = prevHunk.afterStartLine

  for (const line of prevHunk.lines) {
    if (line.type === 'unchanged' || line.type === 'deleted') {
      prevBeforeEnd++
    }
    if (line.type === 'unchanged' || line.type === 'added') {
      prevAfterEnd++
    }
  }
  prevBeforeEnd--
  prevAfterEnd--

  const skippedBefore = currentHunk.beforeStartLine - prevBeforeEnd - 1
  const skippedAfter = currentHunk.afterStartLine - prevAfterEnd - 1

  if (skippedBefore > 0 || skippedAfter > 0) {
    return { before: skippedBefore, after: skippedAfter }
  }

  return null
}
