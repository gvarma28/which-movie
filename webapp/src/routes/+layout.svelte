<script lang="ts">
	import '../app.css';
</script>

<div class="flex min-h-screen flex-col">
	<nav class="bg-gray-800 p-4 text-white">
		<div class="container mx-auto flex items-center justify-between">
			<a href="/" class="text-xl font-bold">Movie Finder</a>
			<div class="space-x-4">
				<a href="/" class="hover:text-gray-300">Home</a>
				<a href="/about" class="hover:text-gray-300">About</a>
			</div>
		</div>
	</nav>

	<main class="container mx-auto flex-grow px-4">
		<slot />
	</main>

	<footer class="mt-auto bg-gray-800 p-4 text-white">
		<div class="container mx-auto text-center">
			<p>© 2025 Movie Finder. Made with ❤️</p>
		</div>
	</footer>
</div>
