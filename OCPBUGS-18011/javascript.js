#+BEGIN_EXPORT html
<style>
  /* Alternating row colors */
  tr:nth-child(even) {
    background-color: #f2f2f2;
  }

  /* Right-align all table cells by default */
  td {
    text-align: right;
  }

  /* Left-align the first table cell */
  td:first-child {
    text-align: left !important;
  }

  /* Right-align all table headers by default */
  th {
    text-align: right;
  }

  /* Left-align the first table header */
  th:first-child {
    text-align: left !important;
  }
</style>

<script>
  document.addEventListener("DOMContentLoaded", function() {
    const tables = document.querySelectorAll("table");

    tables.forEach(table => {
      const headerRow = table.querySelector("tr");
      const columnHeader = headerRow ? headerRow.querySelectorAll("th")[1] : null;
      const isLatencyTable = columnHeader && columnHeader.textContent.includes("latency");

      const rows = table.querySelectorAll("tr");
      rows.forEach((row, index) => {
	if(index === 0) return;

	const cells = row.querySelectorAll("td");
	if(cells.length === 0) return;

	const percentageCell = cells[2];
	const percentage = parseFloat(percentageCell.textContent.trim());

	let baseColor, gradientColor;
	if (isLatencyTable) {
	  baseColor = percentage >= 0 ? 'rgba(255, 200, 200, 0.6)' : 'rgba(200, 255, 200, 0.6)';
	  gradientColor = percentage >= 0 ? 'rgba(255, 0, 0, 1)' : 'rgba(0, 255, 0, 1)';
	} else {
	  baseColor = percentage >= 0 ? 'rgba(200, 255, 200, 0.6)' : 'rgba(255, 200, 200, 0.6)';
	  gradientColor = percentage >= 0 ? 'rgba(0, 255, 0, 1)' : 'rgba(255, 0, 0, 1)';
	}

	const magnitude = Math.abs(percentage);
	const gradient = `linear-gradient(to left, ${gradientColor} 0%, ${gradientColor} ${magnitude}%, ${baseColor} ${magnitude}%, ${baseColor} 100%)`;

	percentageCell.style.background = gradient;
      });
    });
  });
</script>

#+END_EXPORT
