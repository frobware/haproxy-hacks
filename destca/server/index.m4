<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Fetch URLs and Display Results</title>
    <style>
    table {
        width: 100%;
        border-collapse: collapse;
    }
    th, td {
        border: 1px solid black;
        padding: 8px;
        text-align: left;
    }
    th {
        background-color: #f2f2f2;
    }
    </style>
</head>
<body>
    <h1>URL Results</h1>
    <table>
    <thead>
        <tr>
        <th>URL</th>
        <th>Response</th>
        </tr>
    </thead>
    <tbody id="results">
    </tbody>
    </table>
    <script>
    const subdomain = 'SUBDOMAIN';  // this will be replaced by M4
    const urls = [
	'https://scripts-with-key-and-cert-destca.apps.' + subdomain + '/test',
	'https://images-with-key-and-cert-destca.apps.' + subdomain + '/test',
    ];

    async function fetchUrl(url, resultCell) {
        try {
            const response = await fetch(url);
            const data = await response.text();
            resultCell.textContent = data;
        } catch (error) {
            console.error('Error fetching URL:', error);
            resultCell.textContent = error.message;
        }
    }

    function fetchUrls() {
	const resultsElement = document.getElementById('results');
	for (const url of urls) {
            const row = document.createElement('tr');
            const urlCell = document.createElement('td');
            const resultCell = document.createElement('td');
            const link = document.createElement('a');  // create a new <a> element
            link.href = url;  // set its href to the URL
            link.textContent = url;  // set its text to the URL
            urlCell.appendChild(link);  // append the <a> element to the URL cell
            row.appendChild(urlCell);
            row.appendChild(resultCell);
            resultsElement.appendChild(row);
            fetchUrl(url, resultCell);
	}
    }

    fetchUrls();
    </script>
</body>
</html>
