function $(id) {
	return document.getElementById(id);
}

const grapher = (function() {
	const container = $("graph-container")
	const local = {}

	const svg = d3.create("svg")
	container.append(svg.node());

	let width;
	let height;
	let offsetWidth;
	let offsetHeight;

	const resize = () => {
		width = container.clientWidth;
		height = container.clientHeight;
		offsetWidth = width / 2;
		offsetHeight = height / 2;

		svg
			.attr("width", width)
			.attr("height", height)
			.attr("viewBox", [-width / 2, -height / 2, width, height])
	}
	resize()

	window.addEventListener("resize", resize)

	const force = d3.forceSimulation()
		.force("link", d3.forceLink().id(d => d.id).distance(100))
		.force("charge", d3.forceManyBody().strength(-300))
		.force("center", d3.forceCenter(0, 0))
		.force("x",
			d3.forceX(d => {
				let value = 0
				if (d.id == local.start) value = 0.1 * width - offsetWidth;
				else if (d.id == local.end) value = 0.9 * width - offsetWidth;
				return value;
			}).strength(d => {
				return d.id == local.start || d.id == local.end ? 0.005 : 0
			})
		).force("y",
			d3.forceY(d => {
				let value = 0
				if (d.id == local.start) value = 0.1 * height - offsetHeight;
				else if (d.id == local.end) value = 0.9 * height - offsetHeight;
				return value;
			}).strength(d => {
				return d.id == local.start || d.id == local.end ? 0.005 : 0
			})
		);
	const link = force.force("link")

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

		if (local.textNodeDOM) local.textNodeDOM
			.attr("transform", d => `translate(${d.x},${d.y - 10})`);
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

		local.textNodeDOM = local.nodeGroup
			.selectAll("text")
			.data(local.nodes)
			.join("text")
			.attr("r", 5)
			.attr("fill", "white")
			.attr("text-anchor", "middle")
			.attr("opacity", 0.6)
			.text(d => d.id)

		local.linkDOM = local.linksGroup
			.attr("stroke", "white")
			.attr("stroke-opacity", 0.6)
			.selectAll("line")
			.data(local.links)
			.join("line")
			.attr("stroke-width", 1);

		force.alphaTarget(1).restart();
	}

	this.addNode = (name) => {
		if (local.nodes.find(e => e.id == name))
			return
		local.nodes.push({ id: name })
	}

	this.addStartNode = (name) => {
		local.start = name
		this.addNode(name)
	}

	this.addEndNode = (name) => {
		local.end = name
		this.addNode(name)
	}

	this.addLink = (from, to) => {
		local.links.push({ "source": from, "target": to })
	}

	return this
}());

const host = new URL(document.URL).host
const ws = new WebSocket("ws://" + host + "/api")
const state = {
	running: false,
}

ws.addEventListener("error", (e) => {
	console.log(e)
})

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
	/** @type {{ status: "error" | "update" | "started" | "finished", message: any}} */
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
		const pages = data.message
		for (let i = 0; i < pages.length; ++i) {
			const page = pages[i]
			if (i == 0) grapher.addStartNode(page)
			else if (i == pages.length - 1) grapher.addEndNode(page)
			else grapher.addNode(page)
			if (i > 0)
				grapher.addLink(pages[i - 1], page)
			grapher.refreshGraph()
		}
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
