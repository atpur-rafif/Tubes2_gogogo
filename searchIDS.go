package main

// import "log"

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

// harusnya udah benar
func BuatGraf(input *Node, depth int) { // root = depth 0
	if depth != 0 {
		listNewNode := getLinks(input.nilai)
		for i := 0; i < len(listNewNode); i++ {
			// newNode := &Node{
			// 	nilai: listNewNode[i],
			// }
			// buatGraf(newNode, depth)
			input.AddTetangga(listNewNode[i])
			// log.Printf(listNewNode[i])
		}
		for i := 0; i < len(input.tetangga); i++ {
			newDepth := depth - 1
			BuatGraf(input.tetangga[i], newDepth)
		}
	}
}

func DLS(input Node, target string, jumlahArtikel *int, rute *[]string, depth int) bool {
	*rute = append(*rute, input.nilai)
	// log.Printf(input.nilai)
	// log.Printf("DLS ke-%d", depth)
	if input.nilai == target {
		// log.Printf("DLS %d", len(*rute))
		// for i := 0; i < len(*rute); i++ {
		// 	log.Printf("rute %d %s", i+1, (*rute)[i])
		// }
		return true
	} else {
		*jumlahArtikel++
		if depth == 0 {
			*rute = (*rute)[:len(*rute)-1]
			return false
		} else {
			for i := 0; i < len(input.tetangga); i++ {
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
	depth := 0
	for {
		graf := &Node{
			nilai: input,
		}
		BuatGraf(graf, depth)
		if DLS(*graf, target, jumlahArtikel, rute, depth) {
			break
		} else {
			depth += 1
		}
	}
}
