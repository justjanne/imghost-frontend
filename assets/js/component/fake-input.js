const keyListener = (event) => {
    if (event.which === 13) {
        event.preventDefault();
    }
};

function initFakeInput() {
    Array.prototype.slice.call(document.querySelectorAll(".fake-input[contenteditable]:not([data-bound_fake_input])")).forEach(elem => {
        if (elem.dataset["multiline"] === undefined) {
            elem.addEventListener("keypress", keyListener);
        }
        elem.dataset["bound_fake_input"] = "true";
        if (element.innerText.trim() === "") {
            element.innerText = "";
        }
    });
}

initFakeInput();