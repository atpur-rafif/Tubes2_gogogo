package main

type Node struct {
	nilai    string
	tetangga []*Node
}

func (n *Node) AddTetangga(tetangga string) {
	newNode := &Node{
		nilai: tetangga,
	}
	n.tetangga = append(n.tetangga, newNode)
}

func BuatGraf(input *Node, depth int) { // root = depth 0
	if depth != 0 {
		listNewNode := getLinks(input.nilai)
		for i := 0; i < len(listNewNode); i++ {
			// newNode := &Node{
			// 	nilai: listNewNode[i],
			// }
			// buatGraf(newNode, depth)
			input.AddTetangga(listNewNode[i])
		}
		for i := 0; i < len(input.tetangga); i++ {
			newDepth := depth - 1
			BuatGraf(input.tetangga[i], newDepth)
		}
	}
}

func DLS(input Node, target string, jumlahArtikel *int, rute *[]string, depth int) bool {
	if input.nilai == target {
		return true
	} else {
		*jumlahArtikel++
		if depth == 0 {
			return false
		} else {
			for i := 0; i < len(input.tetangga); i++ {
				*rute = append(*rute, input.nilai)
				if DLS(*input.tetangga[i], target, jumlahArtikel, rute, depth-1) {
					return true
				}
			}
			*rute = (*rute)[:len(*rute)-1]
			return false
		}
	}
}

func SearchIDS(input string, target string, jumlahArtikel *int, rute *[]string) {
	graf := &Node{
		nilai: input,
	}
	depth := 0
	for {
		BuatGraf(graf, depth)
		hasil := []string{""}
		if DLS(*graf, target, jumlahArtikel, &hasil, depth) {
			break
		} else {
			depth += 1
		}
	}
}
