package lexer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")
var reIdentifier = regexp.MustCompile(`^[_\d\w]+`)
var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
var reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)

var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

type Lexer struct {
	chunk         string // 源代码
	chunkName     string // 源代码名字
	line          int    // 当前行号
	nextToken     string // 下一个Token
	nextTokenKind int    // 下一个Token的类型
	nextTokenLine int    // 下一个Token的行号
}

// 获取下一个token的类型然后恢复
func (self *Lexer) LookAhead() int {
	if self.nextTokenLine > 0 {
		return self.nextTokenKind
	}
	currentLine := self.line
	line, kind, token := self.NextToken()
	self.line = currentLine
	self.nextTokenLine = line
	self.nextTokenKind = kind
	self.nextToken = token
	return kind
}

// 根据文件名和源代码创建Lexer结构体，并将初始行号设置为1
func NewLexer(chunk, chunkName string) *Lexer {
	return &Lexer{chunk, chunkName, 1, "", 0, 0}
}

// 提取指定类型的token
func (self *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, kind_, token := self.NextToken()
	if kind_ != kind {
		self.error("syntax error near '%s'", token)
	}
	return
}

// 提取标识符
func (self *Lexer) NextIdentifier() (line int, name string) {
	return self.NextTokenOfKind(TOKEN_IDENTIFIER)
}

// 返回行号
func (self *Lexer) Line() int {
	return self.line
}

// 跳过空白字符和注释，返回下一个token
func (self *Lexer) NextToken() (line, kind int, token string) {
	// 查看是否有预读的token
	if self.nextTokenLine > 0 {
		line = self.nextTokenLine
		kind = self.nextTokenKind
		token = self.nextToken
		self.line = self.nextTokenLine
		self.nextTokenLine = 0
		return
	}
	self.skipWhiteSpaces()
	if len(self.chunk) == 0 {
		return self.line, TOKEN_EOF, "EOF"
	}

	switch self.chunk[0] {
	case ';':
		self.next(1)
		return self.line, TOKEN_SEP_SEMI, ";"
	case ',':
		self.next(1)
		return self.line, TOKEN_SEP_COMMA, ","
	case '(':
		self.next(1)
		return self.line, TOKEN_SEP_LPAREN, "("
	case ')':
		self.next(1)
		return self.line, TOKEN_SEP_RPAREN, ")"
	case ']':
		self.next(1)
		return self.line, TOKEN_SEP_RBRACK, "]"
	case '{':
		self.next(1)
		return self.line, TOKEN_SEP_LCURLY, "{"
	case '}':
		self.next(1)
		return self.line, TOKEN_SEP_RCURLY, "}"
	case '+':
		self.next(1)
		return self.line, TOKEN_OP_ADD, "+"
	case '-':
		self.next(1)
		return self.line, TOKEN_OP_MINUS, "-"
	case '*':
		self.next(1)
		return self.line, TOKEN_OP_MUL, "*"
	case '^':
		self.next(1)
		return self.line, TOKEN_OP_POW, "^"
	case '%':
		self.next(1)
		return self.line, TOKEN_OP_MOD, "%"
	case '&':
		self.next(1)
		return self.line, TOKEN_OP_BAND, "&"
	case '|':
		self.next(1)
		return self.line, TOKEN_OP_BOR, "|"
	case '#':
		self.next(1)
		return self.line, TOKEN_OP_LEN, "#"
	case ':':
		if self.test("::") {
			self.next(2)
			return self.line, TOKEN_SEP_LABEL, "::"
		} else {
			self.next(1)
			return self.line, TOKEN_SEP_COLON, ":"
		}
	case '/':
		if self.test("//") {
			self.next(2)
			return self.line, TOKEN_OP_IDIV, "//"
		} else {
			self.next(1)
			return self.line, TOKEN_OP_DIV, "/"
		}
	case '~':
		if self.test("~=") {
			self.next(2)
			return self.line, TOKEN_OP_NE, "~="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_WAVE, "~"
		}
	case '=':
		if self.test("==") {
			self.next(2)
			return self.line, TOKEN_OP_EQ, "=="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_ASSIGN, "="
		}
	case '<':
		if self.test("<<") {
			self.next(2)
			return self.line, TOKEN_OP_SHL, "<<"
		} else if self.test("<=") {
			self.next(2)
			return self.line, TOKEN_OP_LE, "<="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_LT, "<"
		}
	case '>':
		if self.test(">>") {
			self.next(2)
			return self.line, TOKEN_OP_SHR, ">>"
		} else if self.test(">=") {
			self.next(2)
			return self.line, TOKEN_OP_GE, ">="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_GT, ">"
		}
	case '.':
		if self.test("...") {
			self.next(3)
			return self.line, TOKEN_VARARG, "..."
		} else if self.test("..") {
			self.next(2)
			return self.line, TOKEN_OP_CONCAT, ".."
		} else if len(self.chunk) == 1 || !isDigit(self.chunk[1]) {
			self.next(1)
			return self.line, TOKEN_SEP_DOT, "."
		}
	case '[':
		if self.test("[[") || self.test("[=") {
			return self.line, TOKEN_STRING, self.scanLongString()
		} else {
			self.next(1)
			return self.line, TOKEN_SEP_LBRACK, "["
		}
	case '\'', '"':
		return self.line, TOKEN_STRING, self.scanShortString()
	}

	// 数字字面量
	c := self.chunk[0]
	if c == '.' || isDigit(c) {
		token := self.scanNumber()
		return self.line, TOKEN_NUMBER, token
	}
	// 标识符和关键字
	if c == '_' || isLetter(c) {
		token := self.scanIdentifier()
		// 判断是否是关键字
		if kind, found := keywords[token]; found {
			return self.line, kind, token // keyword
		} else {
			return self.line, TOKEN_IDENTIFIER, token
		}
	}

	self.error("unexpected symbol near %q", c)
	return
}

// 判断是否是数字
func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// 判断是否是字母
func isLetter(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// 扫描并返回单词
func (self *Lexer) scanIdentifier() string {
	return self.scan(reIdentifier)
}

// 扫描并返回数字
func (self *Lexer) scanNumber() string {
	return self.scan(reNumber)
}

// 扫描并返回数字或标识符
func (self *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(self.chunk); token != "" {
		self.next(len(token))
		return token
	}
	panic("unreachable")
}

// 跳过空白字符和注释
func (self *Lexer) skipWhiteSpaces() {
	for len(self.chunk) > 0 {
		if self.test("--") {
			self.skipComment()
		} else if self.test("\r\n") || self.test("\n\r") {
			self.next(2)
			self.line++
		} else if isNewLine(self.chunk[0]) {
			self.next(1)
			self.line++
		} else if isWhiteSpace(self.chunk[0]) {
			self.next(1)
		} else {
			break
		}
	}
}

// 判断剩余的源代码是否以某种字符串开头
func (self *Lexer) test(s string) bool {
	return strings.HasPrefix(self.chunk, s)
}

// 跳过n个字符
func (self *Lexer) next(n int) {
	self.chunk = self.chunk[n:]
}

// 判断字符是否是空白字符
func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

// 判断字符是否是换行符
func isNewLine(c byte) bool {
	return c == '\n' || c == '\r'
}

// 跳过注释
func (self *Lexer) skipComment() {
	self.next(2) // 跳过"--"
	if self.test("[") {
		if reOpeningLongBracket.FindString(self.chunk) != "" { // 长注释
			self.scanLongString()
			return
		}
	}
	// 跳过单行注释
	for len(self.chunk) > 0 && !isNewLine(self.chunk[0]) {
		self.next(1)
	}
}

// 扫描并返回一个长字符串
func (self *Lexer) scanLongString() string {
	// 使用正则表达式寻找长字符串的开头
	openingLongBracket := reOpeningLongBracket.FindString(self.chunk)
	if openingLongBracket == "" {
		self.error("invalid long string delimiter near '%s'", self.chunk[0:2])
	}
	// 查找]]
	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIndex := strings.Index(self.chunk, closingLongBracket)
	// 没找到报错
	if closingLongBracketIndex < 0 {
		self.error("unfinished long string or comment")
	}
	// 截取长字符串中间的有效部分
	str := self.chunk[len(openingLongBracket):closingLongBracketIndex]
	self.next(closingLongBracketIndex + len(closingLongBracket))

	// 把所有换行符替换为单个换行符,这样做的目的是，即使长字符串中包含多种不同的换行符（例如，Windows 和 Unix 风格的换行符），也可以将它们统一为单个换行符。
	str = reNewLine.ReplaceAllString(str, "\n")
	// 统计长字符串中的行数
	self.line += strings.Count(str, "\n")

	// 如果字符串以换行符开头，将其删除()
	// 这样做的原因是，长字符串的开始标记 [[ 和结束标记 ]] 通常都会出现在独立的一行，因此长字符串本身不应该包含这一行
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}
	return str
}

// 抛出错误信息
func (self *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	err = fmt.Sprintf("%s:%d: %s", self.chunkName, self.line, err)
	panic(err)
}

// 扫描并返回短字符串
func (self *Lexer) scanShortString() string {
	// 使用正则表达式提取短字符串
	if str := reShortStr.FindString(self.chunk); str != "" {
		self.next(len(str))
		// 去掉两段引号
		str = str[1 : len(str)-1]
		// 替换转义字符
		if strings.Index(str, `\`) >= 0 {
			self.line += len(reNewLine.FindAllString(str, -1))
			str = self.escape(str)
		}
		return str
	}
	self.error("unfinished string")
	return ""
}

// 替换转义字符
func (self *Lexer) escape(str string) string {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			self.error("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd(ASCII 码)
			// 查找转义序列
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF { // 0 <= d <= 255 超过报错
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				self.error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX (十六进制)
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX} (Unicode)
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				self.error("UTF-8 value too large near '%s'", found)
			}
		case 'z': // \z (空白)
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { // todo
				str = str[1:]
			}
			continue
		}
		self.error("invalid escape sequence near '\\%c'", str[1])
	}

	return buf.String()
}
