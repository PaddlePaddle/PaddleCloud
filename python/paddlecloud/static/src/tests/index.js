import { assert } from 'chai';

// jsdom-global is a no-op when browserified;
require('jsdom-global')();
window.assert = assert;

if (!document.getElementById('mocha')) {
    document.body.innerHTML = '<div id="renderContainer"></div>';
}

// these need to stay a require() because imports are hoisted
require('./main');