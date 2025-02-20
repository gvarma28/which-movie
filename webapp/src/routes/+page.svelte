<script lang="ts">
	import type { IsValidUrlReponse, Movie } from '$lib/types';

	let youtubeUrl: string = '';
	let movies: Movie[] = [];
	let error: string | null = null;
	let isLoading: boolean = false;

	async function isValidUrl(url: string): Promise<IsValidUrlReponse> {
		try {
			const cleanUrl = url.split('?')[0];
			const urlObj = new URL(cleanUrl);
			if (urlObj.hostname.includes('youtube.com') || urlObj.hostname.includes('youtu.be')) {
				return {
					isValid: true,
					url: cleanUrl
				};
			}
			return {
				isValid: false,
				url: cleanUrl
			};
		} catch {
			return {
				isValid: false,
				url: ''
			};
		}
	}

	async function doMagic(url: string): Promise<any> {
		const SERVER_URL = import.meta.env.VITE_SERVER_URL;
		const api_url = SERVER_URL + 'magic?url=' + url;
		const response: any = await fetch(api_url, {
			method: 'GET'
		});
		const jsonReponse = await response.json();
		return jsonReponse;
	}

	async function findMovie(): Promise<void> {
		try {
			error = null;
			isLoading = true;
			movies = [];
			const { isValid, url } = await isValidUrl(youtubeUrl);
			if (!isValid) {
				throw new Error('Please enter a valid YouTube URL');
			}

			// const jsonReponse = await doMagic(url);

			// const localMovies: Movie[] = [];
			// jsonReponse.result.split(',').forEach((movieStr: string) => {
			// 	localMovies.push({
			// 		title: movieStr.split('(')[0].trim(),
			// 		year: Number(movieStr.split('(')[1].split(')')[0].trim()),
			// 		description: 'This is where the movie details would appear.'
			// 	});
			// });

			// movies = localMovies;
			movies = [
				{
					title: 'The First Movie',
					year: 2024,
					description: 'This is the first movie that matches the clip.'
				},
				{
					title: 'Another Similar Movie',
					year: 2023,
					description: 'This movie also has similar scenes to the YouTube short.'
				},
				{
					title: 'The First Movie',
					year: 2024,
					description: 'This is the first movie that matches the clip.'
				}
			];
		} catch (e) {
			error = e instanceof Error ? e.message : 'An unknown error occurred';
			movies = [];
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex min-h-[calc(100vh-120px)] flex-col">
	<div
		class={movies.length === 0
			? 'flex min-h-[calc(100vh-120px)] items-center justify-center'
			: 'mt-4'}
	>
		<div class="w-full p-4">
			<div class="mx-auto max-w-4xl text-center">
				<h1 class="mb-8 text-4xl font-bold text-[#493628]">Find Movie From YouTube Short</h1>
				<div class="space-y-4">
					<input
						type="text"
						bind:value={youtubeUrl}
						placeholder="Paste YouTube short URL here"
						class="w-full rounded-lg border border-[#D6C0B3] bg-white p-3 placeholder-[#AB886D]/70 focus:outline-none focus:ring-2 focus:ring-[#AB886D]"
					/>
					<button
						on:click={findMovie}
						disabled={isLoading}
						class="rounded-lg bg-[#AB886D] px-6 py-2 text-white transition-colors hover:bg-[#493628] disabled:cursor-not-allowed disabled:bg-[#D6C0B3]"
					>
						{isLoading ? 'Searching...' : 'Find Movie'}
					</button>
				</div>
				{#if error}
					<div class="mt-4 rounded border border-red-400 bg-red-100 p-3 text-red-700" role="alert">
						{error}
					</div>
				{/if}
			</div>
		</div>
	</div>

	{#if movies.length > 0}
		<div class="flex-1 overflow-y-auto p-4">
			<div class="mx-auto max-w-4xl">
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
					{#each movies as movie}
						<div
							class="rounded-lg border border-[#D6C0B3] bg-white p-6 text-left shadow-lg transition-shadow hover:shadow-xl"
						>
							<h2 class="mb-2 text-2xl font-bold text-[#493628]">{movie.title}</h2>
							<p class="mb-2 text-[#AB886D]">Year: {movie.year}</p>
							<p class="text-[#493628]/90">{movie.description}</p>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>
