<script lang="ts">
	import type { IsValidUrlReponse, MagicResponse, Movie } from '$lib/types';

	let youtubeUrl: string = '';
	let movies: Movie[] = [];
	let error: string | null = null;
	let isLoading: boolean = false;

	async function isValidUrl(url: string): Promise<IsValidUrlReponse> {
		try {
			const cleanUrl = url.split('?')?.[0];
			const urlObj = new URL(cleanUrl);
			if (
				(urlObj.hostname.includes('youtube.com') || urlObj.hostname.includes('youtu.be')) &&
				urlObj.pathname.includes('shorts')
			) {
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
		// test result - avoid api call
		if (import.meta.env?.VITE_TESTING_UI === 'true') {
			return {
				results: [
					{
						movie_name: 'The Big Short',
						year: 2015,
						short_description:
							'A group of investors bet against the housing market before the 2008 financial crisis, discovering the corrupt practices that led to the economic collapse. Their journey reveals the complexities of finance and the human cost of greed.'
					}
				],
				success: true
			};
		}
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
				throw new Error('Please enter a valid YouTube Short URL');
			}

			const jsonReponse: MagicResponse = await doMagic(url);
			if (!jsonReponse.success) {
				throw new Error('The server failed to process the url');
			}
			movies = jsonReponse.results;
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
				<form on:submit={findMovie}>
					<div class="space-y-4">
						<input
							type="text"
							bind:value={youtubeUrl}
							placeholder="Paste YouTube short URL here"
							class="w-full rounded-lg border border-[#D6C0B3] bg-white p-3 placeholder-[#AB886D]/70 focus:outline-none focus:ring-2 focus:ring-[#AB886D]"
						/>
						<button
							disabled={isLoading}
							type="submit"
							class="rounded-lg bg-[#AB886D] px-6 py-2 text-white transition-colors hover:bg-[#493628] disabled:cursor-not-allowed disabled:bg-[#D6C0B3]"
						>
							{isLoading ? 'Processing...' : 'Find Movie'}
						</button>
					</div>
				</form>
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
				<div
					class={movies.length === 1
						? 'flex justify-center'
						: 'grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3'}
				>
					{#each movies as movie}
						<div
							class="rounded-lg border border-[#D6C0B3] bg-white p-6 text-left shadow-lg transition-shadow hover:shadow-xl"
						>
							<h2 class="mb-2 text-2xl font-bold text-[#493628]">{movie.movie_name}</h2>
							<p class="mb-2 text-[#AB886D]">Year: {movie.year}</p>
							<p class="text-[#493628]/90">{movie.short_description}</p>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>
