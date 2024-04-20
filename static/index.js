function $(id) {
	return document.getElementById(id);
}

const ws = new WebSocket("/api")
const state = {
	running: false,
}

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

let timerId = 0
function startTimer() {
	$("time-taken").innerText = "0.0"
	timerId = setInterval(() => {
		const from = $("time-taken").innerText
		$("time-taken").innerText = (parseFloat(from) + 0.1).toFixed(1)
	}, 100)
}

function stopTimer() {
	clearInterval(timerId)
}

ws.addEventListener("message", (e) => {
	/** @type {{ status: "error" | "update" | "started" | "finished", message: string}} */
	const data = JSON.parse(e.data)
	if (data.status == "error") {
		alert(data.message)
		return
	} else if (data.status == "started") {
		state.running = true
		clearUpdate()
		showUpdate("Started...")
		startTimer()
	}
	else if (data.status == "update") showUpdate(data.message)
	else if (data.status == "finished") {
		state.running = false
		clearUpdate()
		showUpdate(data.message)
		stopTimer()
	}
})

$("input-start").value = "Highway"
$("input-end").value = "Traffic"
$("search-button").addEventListener("click", async () => {
	$("search-button").blur()
	let force = false
	if (state.running) {
		if (!confirm("Program still running, cancel and search the new one?")) {
			return
		}
		force = true
	}

	const start = $("input-start").value
	const end = $("input-end").value
	const type = $("input-method").value

	if (start == "" || end == "") {
		alert("Input still empty!")
		return
	}

	ws.send(JSON.stringify({
		start, end, type, force
	}))
})
