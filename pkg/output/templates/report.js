(() => {
	const toolbar = document.querySelector(".toolbar");
	if (!toolbar) return;

	const input = toolbar.querySelector("input");
	const status = toolbar.querySelector(".filter-status");
	const sections = [...document.querySelectorAll(".section")].map((el) => ({
		el,
		count: el.querySelector(".count"),
		copy: el.querySelector(".copy"),
		rows: [...el.querySelectorAll("tbody tr")].map((tr) => ({ tr, text: tr.textContent.toLowerCase() })),
	}));
	const total = sections.reduce((n, s) => n + s.rows.length, 0);

	// Filter every table at once.
	const applyFilter = () => {
		const q = input.value.trim().toLowerCase();
		let shown = 0;

		for (const s of sections) {
			let n = 0;
			for (const r of s.rows) {
				r.tr.hidden = q !== "" && !r.text.includes(q);
				if (!r.tr.hidden) n++;
			}
			shown += n;
			s.count.textContent = q ? n + " / " + s.rows.length : s.rows.length;
			s.el.classList.toggle("no-match", q !== "" && n === 0 && s.rows.length > 0);
		}
		status.textContent = q ? shown + " of " + total + " rows" : "";
	};

	input.addEventListener("input", applyFilter);
	input.addEventListener("keydown", (e) => {
		if (e.key === "Escape") {
			input.value = "";
			applyFilter();
			input.blur();
		}
	});
	document.addEventListener("keydown", (e) => {
		if (e.key === "/" && document.activeElement !== input) {
			e.preventDefault();
			input.focus();
		}
	});

	// Copy the visible rows of a section, one per line, cells separated by tabs.
	const writeClipboard = (text) => {
		if (navigator.clipboard && window.isSecureContext) return navigator.clipboard.writeText(text);

		const area = document.createElement("textarea");
		area.value = text;
		area.style.position = "fixed";
		area.style.opacity = "0";
		document.body.appendChild(area);
		area.select();
		const ok = document.execCommand("copy");
		area.remove();

		return ok ? Promise.resolve() : Promise.reject(new Error("copy failed"));
	};

	for (const s of sections) {
		if (s.rows.length === 0) continue;

		const label = s.copy.querySelector("span");
		s.copy.hidden = false;
		s.copy.addEventListener("click", () => {
			const lines = s.rows
				.filter((r) => !r.tr.hidden)
				.map((r) => [...r.tr.cells].map((td) => td.textContent.replace(/\s+/g, " ").trim()).join("\t"));
			const flash = (text) => {
				label.textContent = text;
				setTimeout(() => (label.textContent = "Copy"), 1500);
			};

			writeClipboard(lines.join("\n")).then(() => flash("Copied " + lines.length), () => flash("Copy failed"));
		});
	}

	// Summary cards jump to (and expand) their section.
	for (const link of document.querySelectorAll("a.stat")) {
		link.addEventListener("click", () => {
			const target = document.querySelector(link.getAttribute("href") + " details");
			if (target) target.open = true;
		});
	}

	window.addEventListener("beforeprint", () => {
		for (const d of document.querySelectorAll("details")) d.open = true;
	});

	toolbar.hidden = false;
})();
