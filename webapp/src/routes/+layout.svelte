<script lang="ts">
	import '../app.css';
</script>

<div class="flex min-h-screen flex-col">
	<nav class="bg-[#493628] p-4 text-white">
		<div class="container mx-auto flex items-center justify-between">
			<a href="/" class="ml-4 text-xl font-bold">Which Movie</a>
			<div class="mr-4 space-x-4">
				<a href="/" class="transition-colors hover:text-[#D6C0B3]">Home</a>
				<a href="/about" class="transition-colors hover:text-[#D6C0B3]">About</a>
			</div>
		</div>
	</nav>

	<main class="flex-1 overflow-hidden">
		<slot />
	</main>

	<footer class="bg-[#493628] p-4 text-white">
		<div class="container mx-auto text-center">
			<p>© 2025 Movie Finder. Made with ❤️</p>
		</div>
	</footer>
</div>
