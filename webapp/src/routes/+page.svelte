<script lang="ts">
	import { extractor } from '../services/extractor/index'
	let url = ''
	let value: any = ['testing', '123'] 

	const onSubmit = async () => {
		console.log('running onSubmit')
		value = await extractor(url);
		console.log('onSubmit finished')
	}
	$: {
		console.log(url);
	}
</script>

<input type="url" bind:value={url}>
<button type="submit" on:click={onSubmit}>submit</button>
<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://svelte.dev/docs/kit">svelte.dev/docs/kit</a> to read the documentation</p>
<p>Dummy {value}</p>
