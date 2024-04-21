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
		stopTimer()
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

const grapher = (function() {
	const width = window.innerWidth;
	const height = window.innerHeight;
	const container = $("graph-container")

	const local = {}

	const svg = d3.create("svg")
		.attr("width", width)
		.attr("height", height)
		.attr("viewBox", [-width / 2, -height / 2, width, height])
	container.append(svg.node());


	const link = d3.forceLink().id(d => d.id)
	const force = d3.forceSimulation()
		.force("link", link)
		.force("charge", d3.forceManyBody().strength(-100))
		.force("center", d3.forceCenter(0, 0))
		.force("x", d3.forceX())
		.force("y", d3.forceY());

	local.nodes = force.nodes()
	local.links = link.links()

	local.nodeGroup = svg.append("g")
	local.linksGroup = svg.append("g")

	force.on("tick", () => {
		if (local.linkDOM) local.linkDOM
			.attr("x1", d => d.source.x)
			.attr("y1", d => d.source.y)
			.attr("x2", d => d.target.x)
			.attr("y2", d => d.target.y);

		if (local.nodeDOM) local.nodeDOM
			.attr("cx", d => d.x)
			.attr("cy", d => d.y);
	});

	this.refreshGraph = () => {
		force.stop()

		force.nodes(local.nodes)
		link.links(local.links)

		local.nodeDOM = local.nodeGroup
			.selectAll("circle")
			.data(local.nodes)
			.join("circle")
			.attr("r", 5)
			.attr("fill", "white")

		local.linkDOM = local.linksGroup
			.attr("stroke", "white")
			.attr("stroke-opacity", 0.6)
			.selectAll("line")
			.data(local.links)
			.join("line")
			.attr("stroke-width", d => Math.sqrt(d.value));

		force.alphaTarget(0.3).restart();
	}

	this.addNode = (name) => {
		local.nodes.push({ id: name })
	}

	this.addLink = (from, to) => {
		local.links.push({ "source": from, "target": to })
	}

	return this
}());

setTimeout(() => {
	for (let i = 0; i < 1000; ++i) {
		grapher.addNode(i)
	}
	grapher.refreshGraph()
}, 1000)
