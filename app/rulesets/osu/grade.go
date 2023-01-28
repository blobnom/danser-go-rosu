package osu

type Grade uint8

const (
	NONE = Grade(iota)
	D
	_C
	B
	A
	S
	SH
	SS
	SSH
)

func (grade Grade) String() string {
	switch grade {
	case D:
		return "D"
	case _C:
		return "C"
	case B:
		return "B"
	case A:
		return "A"
	case S:
		return "S"
	case SH:
		return "SH"
	case SS:
		return "SS"
	case SSH:
		return "SSH"
	case NONE:
		return "None"
	default:
		panic("invalid grade")
	}
}

func (grade Grade) TextureName() string {
	switch grade {
	case D:
		return "d"
	case _C:
		return "c"
	case B:
		return "b"
	case A:
		return "a"
	case S:
		return "s"
	case SH:
		return "sh"
	case SS:
		return "x"
	case SSH:
		return "xh"
	case NONE:
		return "none"
	default:
		panic("invalid grade")
	}
}
