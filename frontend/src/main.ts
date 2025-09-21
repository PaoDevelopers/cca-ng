import { mount } from 'svelte';
import App from './App.svelte';

console.log('£ Main: Application starting');
console.log('£ Main: Looking for app target element');

const appElement = document.getElementById('app');
if (!appElement) {
	console.error(' Main: App element not found in DOM');
	throw new Error('App element not found');
}

console.log('£ Main: App element found, mounting Svelte app');

const app = mount(App, {
	target: appElement
});

console.log('£ Main: Svelte app mounted successfully', { app });

export default app;
