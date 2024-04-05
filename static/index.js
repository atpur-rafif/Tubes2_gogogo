function $(id) {
	return document.getElementById(id);
}

const ws = new WebSocket("/api")

ws.addEventListener("error", (e) => {
	console.log(e)
})

ws.addEventListener("message", (e) => {
	/** @type {{ status: "error" | "update" | "success", message: string}} */
	const data = JSON.parse(e.data)
	console.log(data)
})

$("search-button").addEventListener("click", async () => {
	const start = $("input-start").value
	const end = $("input-end").value

	// if (start == "" || end == "") {
	// 	alert("Input still empty!")
	// 	return
	// }

	ws.send(JSON.stringify({
		start, end
	}))
})
