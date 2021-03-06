package ByteOpt

import "encoding/binary"

/* encode 8 bits unsigned int */
func Encode8u(p []byte, c byte) []byte {
	p[0] = c
	return p[1:]
}

/* decode 8 bits unsigned int */
func Decode8u(p []byte, c *byte) []byte {
	*c = p[0]
	return p[1:]
}

/* encode 16 bits unsigned int (lsb) */
func Encode16u(p []byte, w uint16) []byte {
	binary.BigEndian.PutUint16(p, w)
	return p[2:]
}

/* decode 16 bits unsigned int (lsb) */
func Decode16u(p []byte, w *uint16) []byte {
	*w = binary.BigEndian.Uint16(p)
	return p[2:]
}

/* encode 32 bits unsigned int (lsb) */
func Encode32u(p []byte, l uint32) []byte {
	binary.BigEndian.PutUint32(p, l)
	return p[4:]
}

/* decode 32 bits unsigned int (lsb) */
func Decode32u(p []byte, l *uint32) []byte {
	*l = binary.BigEndian.Uint32(p)
	return p[4:]
}

/* encode 64 bits unsigned int (lsb) */
func Encode64u(p []byte, l uint64) []byte {
	binary.BigEndian.PutUint64(p, l)
	return p[8:]
}

/* decode 64 bits unsigned int (lsb) */
func Decode64u(p []byte, l *uint64) []byte {
	*l = binary.BigEndian.Uint64(p)
	return p[8:]
}

/* encode 64 bits int (lsb) */
func Encode64(p []byte, l int64) []byte {
	binary.BigEndian.PutUint64(p, uint64(l))
	return p[8:]
}

/* decode 64 bits int (lsb) */
func Decode64(p []byte, l *int64) []byte {
	*l = int64(binary.BigEndian.Uint64(p))
	return p[8:]
}
