package crypto

func generateKey(password []byte) []byte {
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		if i < len(password) {
			key[i] = password[i]
		} else {
			key[i] = password[i%len(password)]
		}
	}
	return key
}

func padData(data []byte) []byte {
	padLen := 16 - (len(data) % 16)
	padded := make([]byte, len(data)+padLen)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}
	return padded
}

func unpadData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padLen := int(data[len(data)-1])
	if padLen > 16 || padLen > len(data) {
		return data
	}
	return data[:len(data)-padLen]
}

func xorBlock(dst, a, b []byte) {
	for i := range dst {
		dst[i] = a[i] ^ b[i]
	}
}

func Encrypt(data, key []byte) []byte {
	key = generateKey(key)
	data = padData(data)
	blocks := len(data) / 16
	encrypted := make([]byte, len(data))
	prevBlock := key[:16]

	for i := 0; i < blocks; i++ {
		start := i * 16
		end := start + 16
		block := data[start:end]
		xorBlock(encrypted[start:end], block, prevBlock)
		prevBlock = encrypted[start:end]
	}

	return encrypted
}

func Decrypt(data, key []byte) []byte {
	key = generateKey(key)
	blocks := len(data) / 16
	decrypted := make([]byte, len(data))
	prevBlock := key[:16]

	for i := 0; i < blocks; i++ {
		start := i * 16
		end := start + 16
		block := data[start:end]
		xorBlock(decrypted[start:end], block, prevBlock)
		prevBlock = block
	}

	return unpadData(decrypted)
}
