<script lang="ts">
	import '../app.css';
	let { children } = $props();
</script>

<div>This is a header</div>

{@render children()}
