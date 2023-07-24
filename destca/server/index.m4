<!-- -*- mode: html -*- -->

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
        <th>Time (ms)</th>
        </tr>
    </thead>
    <tbody id="results">
    </tbody>
    </table>
    <script>
    const subdomain = 'SUBDOMAIN';  // this will be replaced by M4
    const namespace = 'NAMESPACE';  // this will be replaced by M4
    const urls = [
	'https://payroll-' + namespace + '.' + subdomain + '/test',
	'https://catpictures-' + namespace + '.' + subdomain + '/test',
    ];
    async function fetchUrl(url) {
	const options = {
            method: 'GET',
            mode: 'cors',
            cache: 'no-cache',
            redirect: 'follow',
            referrerPolicy: 'no-referrer',
	};

        const startMark = 'start-${url}';
        const endMark = 'end-${url}';
        performance.mark(startMark);
        try {
            const response = await fetch(url, options);
            const data = await response.text();
            performance.mark(endMark);
            performance.measure('Fetching: ${url}', startMark, endMark);
            const measure = performance.getEntriesByName('Fetching: ${url}')[0];
            return { data, duration: measure.duration };
        } catch (error) {
            console.error('Error fetching URL:', error);
            return { error: error.message, duration: null };
        }
    }

    async function fetchUrls() {
        const resultsElement = document.getElementById('results');
	for (let i = 0; i < 100; i++) {
            for (const url of urls) {
		const { data, duration } = await fetchUrl(url);
		const row = document.createElement('tr');
		const urlCell = document.createElement('td');
		const resultCell = document.createElement('td');
		const timeCell = document.createElement('td');

		const link = document.createElement('a');  // create a new <a> element
		link.href = url;  // set its href to the URL
		link.textContent = url;  // set its text to the URL
		urlCell.appendChild(link);  // append the <a> element to the URL cell

		resultCell.textContent = data;
		timeCell.textContent = duration ? duration.toFixed(2) + ' ms' : 'Error';

		row.appendChild(urlCell);
		row.appendChild(resultCell);
		row.appendChild(timeCell);

		resultsElement.appendChild(row);
            }
        }
    }

    fetchUrls();
    </script>
</body>
</html>
