import { useMemo } from 'react'
import { computeLineDiff, getSkippedLines, type DiffHunk, type DiffLine } from '@/utils/line-diff'

interface SideBySideDiffProps {
  before: string
  after: string
  contextLines?: number
}

export function SideBySideDiff({
  before,
  after,
  contextLines = 5,
}: SideBySideDiffProps) {
  const hunks = useMemo(
    () => computeLineDiff(before, after, contextLines),
    [before, after, contextLines]
  )

  if (hunks.length === 0) {
    return (
      <div className="rounded border border-gray-200 bg-gray-50 p-3 text-center text-sm text-gray-500">
        No changes
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded border border-gray-200 font-mono text-xs">
      <div className="flex min-w-[600px]">
        {/* Left side (Before) */}
        <div className="flex-1 border-r border-gray-200">
          <div className="sticky top-0 bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600 border-b border-gray-200">
            Before
          </div>
          <div>
            {hunks.map((hunk, hunkIndex) => (
              <HunkLines
                key={hunkIndex}
                hunk={hunk}
                hunkIndex={hunkIndex}
                hunks={hunks}
                side="before"
              />
            ))}
          </div>
        </div>

        {/* Right side (After) */}
        <div className="flex-1">
          <div className="sticky top-0 bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600 border-b border-gray-200">
            After
          </div>
          <div>
            {hunks.map((hunk, hunkIndex) => (
              <HunkLines
                key={hunkIndex}
                hunk={hunk}
                hunkIndex={hunkIndex}
                hunks={hunks}
                side="after"
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

interface HunkLinesProps {
  hunk: DiffHunk
  hunkIndex: number
  hunks: DiffHunk[]
  side: 'before' | 'after'
}

function HunkLines({ hunk, hunkIndex, hunks, side }: HunkLinesProps) {
  const skipped = getSkippedLines(hunks, hunkIndex)

  // 行をside-by-side表示用にペアリング
  const pairedLines = useMemo(() => pairLines(hunk.lines), [hunk.lines])

  return (
    <>
      {/* Skipped lines separator */}
      {skipped && (skipped.before > 0 || skipped.after > 0) && (
        <div className="bg-blue-50 px-2 py-0.5 text-center text-blue-600 border-y border-blue-100">
          ··· {side === 'before' ? skipped.before : skipped.after} lines ···
        </div>
      )}

      {/* Lines */}
      {pairedLines.map((pair, index) => {
        const line = side === 'before' ? pair.before : pair.after

        if (!line) {
          // Empty placeholder for alignment
          return (
            <div
              key={index}
              className="flex bg-gray-50 border-b border-gray-100"
            >
              <span className="w-10 flex-shrink-0 select-none text-right pr-2 py-0.5 text-gray-300 bg-gray-100">
                {' '}
              </span>
              <span className="flex-1 px-2 py-0.5 whitespace-pre">
                {'\u00A0'}
              </span>
            </div>
          )
        }

        const lineNo = side === 'before' ? line.beforeLineNo : line.afterLineNo
        const { bgClass, lineNoClass } = getLineStyles(line.type, side)

        return (
          <div
            key={index}
            className={`flex border-b border-gray-100 ${bgClass}`}
          >
            <span className={`w-10 flex-shrink-0 select-none text-right pr-2 py-0.5 ${lineNoClass}`}>
              {lineNo ?? ' '}
            </span>
            <span className="flex-1 px-2 py-0.5 whitespace-pre overflow-hidden text-ellipsis">
              {line.content || '\u00A0'}
            </span>
          </div>
        )
      })}
    </>
  )
}

interface LinePair {
  before: DiffLine | null
  after: DiffLine | null
}

/**
 * 行をbefore/afterでペアリング
 * - unchanged: 両側に同じ行
 * - deleted: 左のみ、右は空
 * - added: 右のみ、左は空
 * - 連続するdeleted+addedは並べて表示
 */
function pairLines(lines: DiffLine[]): LinePair[] {
  const pairs: LinePair[] = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    if (line.type === 'unchanged') {
      pairs.push({ before: line, after: line })
      i++
    } else if (line.type === 'deleted') {
      // 連続する削除行を収集
      const deletions: DiffLine[] = []
      while (i < lines.length && lines[i].type === 'deleted') {
        deletions.push(lines[i])
        i++
      }

      // 直後の連続する追加行を収集
      const additions: DiffLine[] = []
      while (i < lines.length && lines[i].type === 'added') {
        additions.push(lines[i])
        i++
      }

      // 削除と追加をペアリング
      const maxLen = Math.max(deletions.length, additions.length)
      for (let j = 0; j < maxLen; j++) {
        pairs.push({
          before: deletions[j] ?? null,
          after: additions[j] ?? null,
        })
      }
    } else if (line.type === 'added') {
      // 削除なしの追加
      pairs.push({ before: null, after: line })
      i++
    }
  }

  return pairs
}

function getLineStyles(type: DiffLine['type'], side: 'before' | 'after'): {
  bgClass: string
  lineNoClass: string
} {
  if (type === 'deleted' && side === 'before') {
    return {
      bgClass: 'bg-red-50',
      lineNoClass: 'bg-red-100 text-red-400',
    }
  }
  if (type === 'added' && side === 'after') {
    return {
      bgClass: 'bg-green-50',
      lineNoClass: 'bg-green-100 text-green-400',
    }
  }
  return {
    bgClass: '',
    lineNoClass: 'bg-gray-100 text-gray-400',
  }
}
