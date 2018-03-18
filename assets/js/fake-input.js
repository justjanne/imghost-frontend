const keyListener = (event) => {
    console.log(event.which);
    if (event.which === 13) {
        event.preventDefault();
    }
};

const changeListener = (event) => {
    requestAnimationFrame(() => {
        event.target.innerText = event.target.innerText
    })
};

Array.prototype.slice.call(document.querySelectorAll(".fake-input[contenteditable]")).forEach(elem => {
    elem.addEventListener("keypress", keyListener);
    elem.addEventListener("change", changeListener);
});