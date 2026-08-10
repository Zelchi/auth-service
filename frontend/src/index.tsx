import { render } from 'solid-js/web';
import App from './App';
import { applyAuthTheme } from './theme';

applyAuthTheme();
render(() => <App />, document.getElementById('root') as HTMLElement);
