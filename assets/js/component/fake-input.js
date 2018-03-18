const keyListener = (event) => {
    if (event.which === 13) {
        event.preventDefault();
    }
};

const changeListener = (event) => {
    requestAnimationFrame(() => {
        const element = event.target;

        if (element.innerText === "\n") {
            element.innerText = "";
        }
    })
};

function initFakeInput() {
    Array.prototype.slice.call(document.querySelectorAll(".fake-input[contenteditable]:not([data-bound_fake_input])")).forEach(elem => {
        elem.addEventListener("input", changeListener);
        if (elem.dataset["multiline"] === undefined) {
            elem.addEventListener("keypress", keyListener);
        }
        elem.dataset["bound_fake_input"] = "true";
    });
}

initFakeInput();