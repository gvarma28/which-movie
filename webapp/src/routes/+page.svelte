<script lang="ts">
	let url = '';
	let value: any = '';

	// "https://www.youtube.com/shorts/6Rrb0GohNoY"
	// "https://www.youtube.com/shorts/16tWbpk8sws"

	const onSubmit = async () => {
		value = ""
		const SERVER_URL = import.meta.env.VITE_SERVER_URL;
		const api_url = SERVER_URL + 'magic?url=' + url;
		const response: any = await fetch(api_url, {
			method: 'GET'
		});
		const jsonReponse = await response.json();
		value = `Movie Name: ${jsonReponse.result}`
	};
	/*
		//  runs whenever `url` changes -> similar to useEffect
		$: {
			console.log(url);
		}
	*/
</script>

<div class="page">
	<div class="container">
		<input type="url" placeholder="Enter YouTube Shorts URL" bind:value={url} />
		<button type="submit" on:click={onSubmit}>submit</button>
		<p>{value}</p>
	</div>
</div>

<style>
	/* Basic reset for margins and padding */
	* {
		margin: 0;
		padding: 0;
		box-sizing: border-box;
	}

	/* Flexbox centering for the container */
	.page {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 100vh; /* Ensure the page takes full height */
		background-color: #f4f7fc;
		padding: 20px;
	}

	/* Container styling */
	.container {
		background-color: white;
		border-radius: 8px;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
		padding: 20px;
		width: 100%;
		max-width: 400px;
		text-align: center;
	}

	/* Input and button styling */
	input[type='url'] {
		width: 100%;
		padding: 10px;
		border: 1px solid #ddd;
		border-radius: 4px;
		margin-bottom: 20px;
		font-size: 1rem;
	}

	button {
		padding: 10px 20px;
		background-color: #4a90e2;
		color: white;
		border: none;
		border-radius: 4px;
		font-size: 1rem;
		cursor: pointer;
		transition: background-color 0.3s ease;
	}

	button:hover {
		background-color: #357ab7;
	}

	/* Result display */
	p {
		font-size: 1.2rem;
		color: #555;
		margin-top: 20px;
	}
</style>
