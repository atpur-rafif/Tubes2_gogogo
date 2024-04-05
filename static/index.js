function $(id) {
	return document.getElementById(id);
}

$("search-button").addEventListener("click", async () => {
	const start = $("input-start").value
	const end = $("input-end").value

	if (start == "" || end == "") {
		alert("Input still empty!")
		return
	}

	const req = await fetch(encodeURI(`/api?start=${start}&end=${end}`))
	const body = await req.text()
	console.log(body)
})
