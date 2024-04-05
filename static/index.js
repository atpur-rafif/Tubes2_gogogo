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

	console.log(start, end)
})
