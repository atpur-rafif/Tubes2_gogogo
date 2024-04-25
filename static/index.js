function $(id) {
	return document.getElementById(id);
}

const host = new URL(document.URL).host
const ws = new WebSocket("ws://" + host + "/api")
const state = {
	running: false,
}

ws.addEventListener("error", (e) => {
	console.log(e)
})

function updateLog(str) {

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
	}
	else if (data.status == "update") {
		$("log-container").innerHTML = data.message.replaceAll("\n", "<br>")
	}
	else if (data.status == "found") {
		$("result-container").insertAdjacentHTML("beforeend", "<p>" + data.message + "</p>")
	}
	else if (data.status == "finished") {
		state.running = false
	}
})

$("input-start").value = "Hitler"
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
