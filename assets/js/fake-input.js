const keyListener = (event) => {
    console.log(event.which);
    if (event.which === 13) {
        event.preventDefault();
    }
};

Array.prototype.slice.call(document.querySelectorAll(".fake-input[contenteditable]")).forEach(elem => {
    elem.addEventListener("keypress", keyListener)
});