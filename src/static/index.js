function $(id) {
	return document.getElementById(id);
}

$("input-start").value = "Medan Prijaji";
$("input-end").value = "Hitler";

// setTimeout(() => {
// 	$("search-button").click()
// }, 100)

const grapher = (function () {
	const local = {
		path: [],
		relatedLink: {},
		relatedNode: {},
		relatedPathAndTime: {},
		pathDepth: {},
		start: null,
		end: null,
		selected: null,
		selectionPriority: {},
	};
	const container = $("graph-container");

	const zoom = d3.zoom().on("zoom", (e) => {
		d3.selectAll("g").attr("transform", e.transform);
	});

	const svg = d3
		.create("svg")
		.call(zoom)
		.on("click", function () {
			clearSelection(2);
		});
	container.append(svg.node());

	let width;
	let height;
	let offsetWidth;
	let offsetHeight;

	const resize = () => {
		width = container.clientWidth;
		height = container.clientHeight;
		offsetWidth = -width / 2;
		offsetHeight = -height / 2;

		svg
			.attr("width", width)
			.attr("height", height)
			.attr("viewBox", [-width / 2, -height / 2, width, height]);
	};
	resize();

	window.addEventListener("resize", resize);

	// Too lazy to choose color https://stackoverflow.com/a/41849519
	const colors = d3
		.scaleLinear()
		.range([
			"#5E4FA2",
			"#3288BD",
			"#66C2A5",
			"#ABDDA4",
			"#E6F598",
			"#FFFFBF",
			"#FEE08B",
			"#FDAE61",
			"#F46D43",
			"#D53E4F",
			"#9E0142",
		]);

	const force = d3
		.forceSimulation()
		.force(
			"link",
			d3
				.forceLink()
				.id((d) => d.id)
				.distance(100),
		)
		.force("charge", d3.forceManyBody().strength(-300))
		.force("center", d3.forceCenter(0, 0))
		.force(
			"x",
			d3
				.forceX((d) => {
					let value = 0;
					if (d.id == local.start) value = 0.1 * width + offsetWidth;
					else if (d.id == local.end) value = 0.9 * width + offsetWidth;
					return value;
				})
				.strength((d) => {
					return d.id == local.start || d.id == local.end ? 0.1 : 0;
				}),
		)
		.force(
			"y",
			d3
				.forceY((d) => {
					let value = 0;
					if (d.id == local.start) value = 0.1 * height + offsetHeight;
					else if (d.id == local.end) value = 0.9 * height + offsetHeight;
					return value;
				})
				.strength((d) => {
					return d.id == local.start || d.id == local.end ? 0.1 : 0;
				}),
		);
	const link = force.force("link");

	local.nodes = force.nodes();
	local.links = link.links();

	local.linksGroup = svg.append("g");
	local.textNodeGroup = svg.append("g");
	local.nodeGroup = svg.append("g").attr("cursor", "pointer");

	force.on("tick", () => {
		if (local.linkDOM)
			local.linkDOM
				.attr("x1", (d) => d.source.x)
				.attr("y1", (d) => d.source.y)
				.attr("x2", (d) => d.target.x)
				.attr("y2", (d) => d.target.y);

		if (local.nodeDOM)
			local.nodeDOM.attr("cx", (d) => d.x).attr("cy", (d) => d.y);

		if (local.textNodeDOM)
			local.textNodeDOM.attr(
				"transform",
				(d) => `translate(${d.x},${d.y - 10})`,
			);
	});

	const refreshSelectionPriority = () => {
		let max = -1;
		Object.entries(local.selectionPriority).forEach(([k, v]) => {
			const vk = parseInt(k);
			if (vk > max) {
				max = vk;
				local.selected = v;
			}
		});
		if (max == -1) local.selected = null;
	};

	const refreshSelection = () => {
		const paths = local.relatedPathAndTime[local.selected];
		const infoDom = $("graph-info-container");

		if (local.selected == null) {
			$("graph-info-container").innerHTML = "";
			return;
		}

		infoDom.innerHTML = "";
		let header = "";
		if (local.selected == local.selectionPriority[2]) header += "🔒 ";
		header += `${local.pathDepth[local.selected]} <a href="https://en.wikipedia.org/wiki/${local.selected}">${local.selected}</a> (${paths.length}):`;
		infoDom.insertAdjacentHTML("beforeend", `<p>${header}</p>`);
		infoDom.insertAdjacentHTML(
			"beforeend",
			paths
				.map(
					(v, i) =>
						`<p>${i + 1}. ${v[0].join(" ➡️ ")} @ ${(v[1] / 1e3).toFixed(3)}s</p>`,
				)
				.join(""),
		);
	};

	const setSelection = (select, priority) => {
		local.selectionPriority[priority] = select;
		refreshSelectionPriority();
		refreshGraph();
	};

	const clearSelection = (priority) => {
		delete local.selectionPriority[priority];
		refreshSelectionPriority();
		refreshGraph();
	};

	const refreshGraph = () => {
		force.stop();
		refreshSelection();

		force.nodes(local.nodes);
		link.links(local.links);

		local.nodeDOM = local.nodeGroup
			.selectAll("circle")
			.data(local.nodes)
			.join("circle")
			.attr("r", (d) => (d.id == local.selectionPriority[2] ? 9 : 7))
			.attr("fill", (d) =>
				d.id == local.selectionPriority[2]
					? "white"
					: colors(local.pathDepth[d.id]),
			)
			.attr("opacity", (d) => {
				const id = d.id;
				if (local.selected != null) {
					if (id == local.selected) return 1;
					if (local.relatedNode[local.selected][id]) return 0.6;
					return 0.2;
				}
				return 0.9;
			})
			.on("click", function (e) {
				const id = this.__data__.id;
				if (local.selectionPriority[2] == id) clearSelection(2);
				else setSelection(this.__data__.id, 2);
				e.stopPropagation();
				refreshGraph();
			})
			.on("pointerover", function () {
				setSelection(this.__data__.id, 0);
			})
			.on("pointerout", function () {
				clearSelection(0);
			})
			.call(
				d3
					.drag()
					.on("start", function (e) {
						setSelection(this.__data__.id, 1);
						if (!e.active) force.alphaTarget(0.3).restart();
						e.subject.fx = e.subject.x;
						e.subject.fy = e.subject.y;
					})
					.on("drag", function (e) {
						e.subject.fx = e.x;
						e.subject.fy = e.y;
					})
					.on("end", function (e) {
						clearSelection(1);
						if (!e.active) force.alphaTarget(0);
						e.subject.fx = null;
						e.subject.fy = null;
					}),
			);

		local.textNodeDOM = local.textNodeGroup
			.selectAll("text")
			.data(local.nodes)
			.join("text")
			.attr("fill", "white")
			.attr("text-anchor", "middle")
			.attr("opacity", (d) => {
				const id = d.id;
				if (local.selected != null) {
					if (id == local.selected) return 1;
					if (local.relatedNode[local.selected][id]) return 0.4;
					return 0.1;
				}
				return 0.6;
			})
			.text((d) => d.id);

		local.linkDOM = local.linksGroup
			.attr("stroke", "white")
			.selectAll("line")
			.data(local.links)
			.join("line")
			.attr("stroke-opacity", (d) => {
				if (local.selected != null) {
					const key = d.source.id + "-" + d.target.id;
					if (local.relatedLink[local.selected][key]) return 0.8;
					return 0.1;
				}
				return 0.6;
			})
			.attr("stroke-width", 1);

		force.alphaTarget(1).restart();
	};
	this.refreshGraph = refreshGraph;

	this.reset = () => {
		local.links = [];
		local.nodes = [];
		local.path = [];
		local.relatedLink = {};
		local.relatedNode = {};
		local.relatedPathAndTime = {};
		local.start = null;
		local.end = null;
		local.selected = null;
		refreshGraph();
	};

	this.addPath = (path, time) => {
		if (local.start == null || local.end == null) {
			local.start = path[0];
			local.end = path[path.length - 1];
			colors.domain(d3.range(path.length));
		} else {
			if (local.start != path[0] || local.end != path[path.length - 1])
				throw "Mismatch start and endpoints with existing path, reset graph to use another endpoints";
		}

		const pathAndTime = [path, time];
		const relatedLink = {};
		for (let i = 0; i < path.length; ++i) {
			const page = path[i];

			if (!local.relatedPathAndTime[page]) local.relatedPathAndTime[page] = [];
			local.relatedPathAndTime[page].push(pathAndTime);

			if (!local.nodes.find((e) => e.id == page)) {
				const newNode = { id: page };
				if (i == 0)
					Object.assign(newNode, {
						x: 0.1 * width + offsetWidth,
						y: 0.1 * height + offsetHeight,
					});
				else if (i == path.length - 1)
					Object.assign(newNode, {
						id: page,
						x: 0.9 * width + offsetWidth,
						y: 0.9 * height + offsetHeight,
					});
				local.nodes.push(newNode);
				local.pathDepth[page] = i;
			}

			if (!local.relatedNode[page]) local.relatedNode[page] = {};
			for (const node of path) local.relatedNode[page][node] = true;

			if (i == 0) continue;

			const from = path[i - 1];
			const to = path[i];

			if (!relatedLink) relatedLink = {};
			relatedLink[from + "-" + to] = true;

			if (!local.links.find((v) => v.source.id == from && v.target.id == to))
				local.links.push({ source: from, target: to });
		}

		for (const node of path) {
			if (!local.relatedLink[node]) local.relatedLink[node] = {};
			Object.assign(local.relatedLink[node], relatedLink);
		}
	};

	return this;
})();

const state = {
	running: false,
	timerId: 0,
	start: 0,
};

function startTimer() {
	$("time-taken").innerText = "0.0";
	state.start = performance.now();
	state.timerId = setInterval(() => {
		const now = performance.now();
		$("time-taken").innerText = ((now - state.start) / 1e3).toFixed(1);
	}, 100);
}

function getTime() {
	const now = performance.now();
	return now - state.start;
}

function stopTimer() {
	clearInterval(state.timerId);
}

function domOnStart() {
	$("search-button").innerText = "Stop";
	$("input-start").disabled =
		$("input-end").disabled =
		$("input-method").disabled =
			true;
	grapher.reset();
}

function domOnFinish() {
	$("search-button").innerText = "Start";
	$("input-start").disabled =
		$("input-end").disabled =
		$("input-method").disabled =
			false;
}

function changeLog(str) {
	$("log-container").innerHTML = str.replaceAll("\n", "<br>");
}

const searchButton = $("search-button");
searchButton.addEventListener("click", async () => {
	searchButton.blur();

	if (state.running) {
		if (!confirm("Program still running, stop searching?")) return;

		state.running = false;
		domOnFinish();
		stopTimer();
		$("time-taken").innerText = "0.0";
		changeLog(`Search stopped at ${getTime() / 1e3}s`);
		ws.send(
			JSON.stringify({
				cancel: true,
			}),
		);
	} else {
		const start = $("input-start").value;
		const end = $("input-end").value;
		const type = $("input-method").value;

		if (start == "" || end == "") {
			alert("Input still empty!");
			return;
		}

		domOnStart();
		ws.send(
			JSON.stringify({
				start,
				end,
				type,
			}),
		);
	}
});

const host = new URL(document.URL).host;
const ws = new WebSocket(
	`ws${document.location.protocol === "https:" ? "s" : ""}://` + host + "/api",
);

ws.addEventListener("error", (e) => {
	console.log(e);
});

ws.addEventListener("close", () => {
	alert("Disconnected with server");
});

ws.addEventListener("message", (e) => {
	/** @type {{ status: "error" | "update" | "started" | "finished", message: any}} */
	const data = JSON.parse(e.data);

	if (data.status == "error") {
		alert(data.message);
		changeLog("Error searching");
		domOnFinish();
	} else if (data.status == "started") {
		state.running = true;
		changeLog(data.message);
		startTimer();
	} else if (data.status == "update") {
		changeLog(data.message);
	} else if (data.status == "found") {
		grapher.addPath(data.message, getTime());
		grapher.refreshGraph();
	} else if (data.status == "finished") {
		state.running = false;
		changeLog(data.message);
		domOnFinish();
		stopTimer();
	}
});
