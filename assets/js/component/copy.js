function initCopy() {
    Array.prototype.slice.call(document.querySelectorAll("button.copy[data-target]:not([data-bound_copy])")).forEach((button) => {
        const target = document.querySelector(button.dataset["target"]);
        if (target) {
            button.addEventListener("click", () => {
                target.select();
                document.execCommand("Copy");
            });
            button.dataset["bound_copy"] = "true";
        }
    });
}
initCopy();