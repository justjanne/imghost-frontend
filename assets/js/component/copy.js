function initCopy() {
    Array.prototype.slice.call(document.querySelectorAll("button.copy[data-target]:not([data-bound-copy])")).forEach((button) => {
        const target = document.querySelector(button.dataset["target"]);
        if (target) {
            button.addEventListener("click", () => {
                target.select();
                document.execCommand("Copy");
            });
            elem.dataset["bound-copy"] = "true";
        }
    });
}
initCopy();