package hosts

const (
	StateStart = iota
	StateWhiteSpace
	StateIP
	StateHostName
	StateComment
)

// ParseLine split a hosts line to three parts.
func ParseLine(line string) (string, []string, string) {
	var ip, comment string
	hosts := make([]string, 0)

	state := StateStart
	start := 0

Outerloop:
	for i, c := range line {
		switch state {
		case StateStart:
			switch c {
			case ';', '#':
				state = StateComment
			default:
				state = StateIP
			}
		case StateWhiteSpace:
			switch c {
			case ';', '#':
				state = StateComment
			case ' ', '\t':
				continue
			default:
				state = StateHostName
				start = i
			}
		case StateIP:
			switch c {
			case '\t', ' ':
				state = StateWhiteSpace
				ip = line[:i]
			}
		case StateHostName:
			switch c {
			case '\t', ' ':
				state = StateWhiteSpace
				hosts = append(hosts, line[start:i])
			}
		case StateComment:
			comment = line[i:]
			break Outerloop
		}
	}

	if state == StateHostName {
		hosts = append(hosts, line[start:])
	}

	return ip, hosts, comment
}
