package hosts

const (
	StateStart = iota
	StateWhiteSpace
	StateIP
	StateHostName
	StateComment
)

// ParseLine split a hosts line to three parts.
// TODO support multiple hostnames in one line.
func ParseLine(line string) (string, string, string) {
	var ip, host, comment string

	state := StateStart
	start := 0

Outerloop:
	for i, c := range line {
		switch state {
		case StateStart:
			switch c {
			case ';':
				fallthrough
			case '#':
				state = StateComment
			default:
				state = StateIP
			}
		case StateWhiteSpace:
			switch c {
			case ';':
				fallthrough
			case '#':
				state = StateComment
			case ' ':
				fallthrough
			case '\t':
				continue
			default:
				state = StateHostName
				start = i
			}
		case StateIP:
			switch c {
			case '\t':
				fallthrough
			case ' ':
				state = StateWhiteSpace
				ip = line[:i]
			}
		case StateHostName:
			switch c {
			case '\t':
				fallthrough
			case ' ':
				state = StateWhiteSpace
				host = line[start:i]
			}
		case StateComment:
			comment = line[i:]
			break Outerloop
		}
	}

	if state == StateHostName {
		host = line[start:]
	}

	return ip, host, comment
}
