function $(id) {
	return document.getElementById(id);
}

const ws = new WebSocket("/api")

ws.addEventListener("error", (e) => {
	console.log(e)
})

function showUpdate(str) {
	const el = $("update-container")
	el.classList.add("show")
	el.insertAdjacentHTML("beforeend", `<p>${str}</p>`)
	el.scrollTop = el.scrollHeight
	if (el.childElementCount > 20) {
		el.firstChild.remove()
	}
}

function clearUpdate() {
	const el = $("update-container")
	el.classList.remove("show")
	el.innerHTML = ""
}

ws.addEventListener("message", (e) => {
	/** @type {{ status: "error" | "update" | "started" | "finished", message: string}} */
	const data = JSON.parse(e.data)
	if (data.status == "error") {
		alert(data.message)
		return
	} else if (data.status == "started") showUpdate("Started...")
	else if (data.status == "update") showUpdate(data.message)
	else if (data.status == "finished") {
		clearUpdate()
		showUpdate(data.message)
	}
})

$("input-start").value = "Highway"
$("input-end").value = "Traffic"
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
