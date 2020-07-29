package editdist

func min(a int, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func levenshtein(a []byte, b []byte) [][]int {
	dist := make([][]int, len(a)+1);
	dist[0] = make([]int, len(b)+1);
	for j := 0; j < len(b)+1; j++ {
		dist[0][j] = j // First row
	}
	for i := 1; i < len(a)+1; i++ {
		dist[i] = make([]int, len(b)+1);
		dist[i][0] = i // First col
		for j := 1; j < len(b)+1; j++ {
			del := dist[i][j-1]+1
			ins := dist[i-1][j]+1
			sub := dist[i-1][j-1]
			if a[i-1] != b[j-1] {
				sub += 1
			}
			dist[i][j] = min(sub, min(del, ins))
		}
	}
	return dist
}

func Levenshtein(a []byte, b []byte) int {
	return levenshtein(a, b)[len(a)][len(b)]
}
