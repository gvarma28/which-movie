<!-- <script lang="ts">
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
</style> -->

<script lang="ts">
	import type { Movie } from '$lib/types';
	import { onMount } from 'svelte';

	let youtubeUrl: string = '';
	let movie: Movie | null = null;
	let error: string | null = null;
	let isLoading: boolean = false;

	async function isValidYoutubeUrl(url: string): Promise<boolean> {
		try {
			const urlObj = new URL(url);
			return urlObj.hostname.includes('youtube.com') || urlObj.hostname.includes('youtu.be');
		} catch {
			return false;
		}
	}

	async function findMovie(): Promise<void> {
		try {
			error = null;
			isLoading = true;

			if (!(await isValidYoutubeUrl(youtubeUrl))) {
				throw new Error('Please enter a valid YouTube URL');
			}

			await new Promise((resolve) => setTimeout(resolve, 1000));

			movie = {
				title: 'Example Movie',
				year: 2024,
				description: 'This is where the movie details would appear.'
			};
		} catch (e) {
			error = e instanceof Error ? e.message : 'An unknown error occurred';
			movie = null;
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex h-[calc(100vh-8rem)] items-center justify-center px-4">
	<div class="w-full max-w-2xl space-y-8">
		<div class="text-center">
			<h1 class="mb-8 text-4xl font-bold">Find Movie from YouTube Short</h1>

			<div class="space-y-4">
				<input
					type="text"
					bind:value={youtubeUrl}
					placeholder="Paste YouTube short URL here"
					class="w-full rounded-lg border p-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
				<button
					on:click={findMovie}
					disabled={isLoading}
					class="rounded-lg bg-blue-600 px-6 py-2 text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-400"
				>
					{isLoading ? 'Searching...' : 'Find Movie'}
				</button>
			</div>

			{#if error}
				<div
					class="mt-4 rounded border border-red-400 bg-red-100 px-4 py-3 text-red-700"
					role="alert"
				>
					{error}
				</div>
			{/if}

			{#if movie}
				<div class="mt-4 rounded-lg bg-white p-6 text-left shadow-lg">
					<h2 class="mb-2 text-2xl font-bold">{movie.title}</h2>
					<p class="mb-2 text-gray-600">Year: {movie.year}</p>
					<p>{movie.description}</p>
				</div>
			{/if}
		</div>
	</div>
</div>
