const keyListener = (event) => {
    console.log(event.which);
    if (event.which === 13) {
        event.preventDefault();
    }
};

const changeListener = (event) => {
    requestAnimationFrame(() => {
        const element = event.target;
        const selectionStart = element.selectionStart;
        const selectionEnd = element.selectionEnd;
        const text = element.innerText;
        element.innerText = (text === "\n") ? "" : text;
        element.selectionStart = selectionStart;
        element.selectionEnd = selectionEnd;
    })
};

Array.prototype.slice.call(document.querySelectorAll(".fake-input[contenteditable]")).forEach(elem => {
    elem.addEventListener("keypress", keyListener);
    elem.addEventListener("input", changeListener);
});