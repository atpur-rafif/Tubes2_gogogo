# WikiRace with BFS & IDS
Program merupakan permanfaatan dari algoritma BFS dan IDS untuk membuat permainan WikiRace

# Algoritma
Program ini menggunakan BFS (Breadth Depth Search) dan IDS (Iterative Deepening Search)

Algoritma BFS menggunakan tipe data queue sehingga node dilalui secara FIFO (First In First Out), sehingga child node pertama akan dilalui lalu dilanjut dengan child node disebelahnya, hingga ditemukan end target

Sementara, algoritma IDS merupakan pengembangan dari DFS, sehingga menggunakan tipe data stack yang dimana node dilalui secara LIFO (Last In First Out), sehingga node akan dilalui hingga kedalaman tertentu, lalu akan dibacktrack dan begitu seterusnya, jika solusi masih belum ditemukan, maka nilai kedalaman akan ditambah

# Requirements
- [Golang](https://go.dev/doc/install)
- [Docker desktop](https://www.docker.com/products/docker-desktop/)
  
# Running
1. Clone repository
   ```
   git clone https://github.com/atpur-rafif/Tubes2_gogogo.git
   ```
2. Pindah direktori
   ```
   cd Tubes2_gogogo/src
   ```
3. Masukkan command
   ```
   go run .
   ```
   atau jika menggunakan docker
   ```
   docker compose up
   ```
4. Buka Peramban
5. Masukkan tautan
   ```
   http://localhost:3000/
   ```
6. Masukkan laman awal dan laman akhir, untuk laman dengan judul memiliki spasi gunakan _
7. Pilih metode pencarian BFS atau IDS
8. Klik tombol _search_

# Author
Group name: gogogo

|Nama	                |NIM
|---------------------|--------
|Haikal Assyauqi	    |13522052
|Benjamin Sihombing	  |13522054
|Muhammad Atpur Rafif	|13522086
