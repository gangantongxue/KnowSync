package chunker

import (
	"math"
	"strings"
)

// Chunk 文档块
type Chunk struct {
	Text  string `json:"text"`
	Index int    `json:"index"`
}

const (
	defaultMaxTokens = 1000
	// 估算：中文约 1.5 字符/token，英文约 4 字符/token
	// 保守使用 2 字符/token
	charsPerToken = 2
)

// Chunker 语义切分器
type Chunker struct {
	maxTokens int
}

// NewChunker 创建语义切分器
func NewChunker(maxTokens int) *Chunker {
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	return &Chunker{maxTokens: maxTokens}
}

// Split 将文章内容按语义边界切分为多个块
func (c *Chunker) Split(content string) []Chunk {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	// 一级分割：按 Markdown 标题
	sections := splitByHeadings(content)

	var chunks []Chunk
	for _, section := range sections {
		if section == "" {
			continue
		}

		// 如果整个 section 在限制内，直接作为一个块
		if estimateTokens(section) <= c.maxTokens {
			chunks = append(chunks, Chunk{Text: section})
			continue
		}

		// 二级分割：按段落（空行）
		subChunks := splitByParagraphs(section, c.maxTokens)
		chunks = append(chunks, subChunks...)
	}

	// 分配索引
	for i := range chunks {
		chunks[i].Index = i
	}

	return chunks
}

// splitByHeadings 按 Markdown 标题分割（# ## ### #### 等）
func splitByHeadings(content string) []string {
	lines := strings.Split(content, "\n")

	var sections []string
	var current strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 检测 Markdown 标题
		if isHeading(trimmed) {
			if current.Len() > 0 {
				sections = append(sections, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
		current.WriteString(line)
		current.WriteString("\n")
	}

	if current.Len() > 0 {
		sections = append(sections, strings.TrimSpace(current.String()))
	}

	return sections
}

// isHeading 判断是否为 Markdown 标题
func isHeading(line string) bool {
	if len(line) == 0 {
		return false
	}
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	return i > 0 && i <= 6 && (len(line) == i || line[i] == ' ')
}

// splitByParagraphs 按空行分割段落，超长段落实行递归切分
func splitByParagraphs(section string, maxTokens int) []Chunk {
	paragraphs := strings.Split(section, "\n\n")

	var chunks []Chunk
	var current strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果加上当前段落会超过限制，先把当前积累的作为一个块
		if current.Len() > 0 && estimateTokens(current.String()+para) > maxTokens {
			chunks = append(chunks, Chunk{Text: strings.TrimSpace(current.String())})
			current.Reset()
		}

		// 段落自身超长，递归切分
		if estimateTokens(para) > maxTokens {
			// 先保存当前积累的内容
			if current.Len() > 0 {
				chunks = append(chunks, Chunk{Text: strings.TrimSpace(current.String())})
				current.Reset()
			}
			// 递归切分长段落
			subChunks := recursiveSplit(para, maxTokens)
			chunks = append(chunks, subChunks...)
			continue
		}

		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(para)
	}

	if current.Len() > 0 {
		chunks = append(chunks, Chunk{Text: strings.TrimSpace(current.String())})
	}

	return chunks
}

// recursiveSplit 递归切分超长文本（按句号、逗号、换行）
func recursiveSplit(text string, maxTokens int) []Chunk {
	if estimateTokens(text) <= maxTokens {
		return []Chunk{{Text: text}}
	}

	// 优先按句号分割
	delimiters := []string{"。", ". ", "！", "？", "\n"}
	for _, delim := range delimiters {
		parts := strings.Split(text, delim)
		if len(parts) <= 1 {
			continue
		}

		var chunks []Chunk
		var current strings.Builder
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			candidate := part
			if current.Len() > 0 {
				candidate = current.String() + delim + part
			}

			if estimateTokens(candidate) <= maxTokens {
				if current.Len() > 0 {
					current.WriteString(delim)
				}
				current.WriteString(part)
			} else {
				if current.Len() > 0 {
					chunks = append(chunks, Chunk{Text: strings.TrimSpace(current.String())})
					current.Reset()
				}
				// 如果单独的部分还是超长，递归切分
				if estimateTokens(part) > maxTokens {
					subChunks := recursiveSplit(part, maxTokens)
					chunks = append(chunks, subChunks...)
				} else {
					current.WriteString(part)
				}
			}

			// 剩余部分
			if i == len(parts)-1 && current.Len() > 0 {
				if i > 0 {
					current.WriteString(delim)
				}
			}
		}

		if current.Len() > 0 {
			chunks = append(chunks, Chunk{Text: strings.TrimSpace(current.String())})
		}

		if len(chunks) > 0 {
			return chunks
		}
	}

	// 兜底：按字符切分
	return splitByChars(text, maxTokens)
}

// splitByChars 按字符数切分（兜底方案）
func splitByChars(text string, maxTokens int) []Chunk {
	maxChars := maxTokens * charsPerToken
	runes := []rune(text)

	var chunks []Chunk
	for i := 0; i < len(runes); i += maxChars {
		end := min(i+maxChars, len(runes))
		chunks = append(chunks, Chunk{Text: string(runes[i:end])})
	}

	return chunks
}

// estimateTokens 估算文本 token 数量
func estimateTokens(text string) int {
	return int(math.Ceil(float64(len([]rune(text))) / float64(charsPerToken)))
}
