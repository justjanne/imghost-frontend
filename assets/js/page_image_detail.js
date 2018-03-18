const fakeTitle = document.querySelector(".title.fake-input[contenteditable]");
const fakeDescription = document.querySelector(".description.fake-input[contenteditable]");

const actualTitle = document.querySelector(".update-form input[name=title]");
const actualDescription = document.querySelector(".update-form input[name=description]");

const fakeTitleListener = (event) => {
    requestAnimationFrame(() => {
        document.title = event.target.innerText + " | i.k8r";
        actualTitle.value = fakeTitle.innerText;
    })

};
const fakeDescriptionListener = (event) => {
    requestAnimationFrame(() => {
        actualDescription.value = fakeDescription.innerText;
    })

};

// Insert <br> between lines instead of \n for editing
fakeDescription.innerHTML = "";
actualDescription.value.split("\n").forEach((line) => {
    const textNode = document.createTextNode(line);
    const brNode = document.createElement("br");
    fakeDescription.appendChild(textNode);
    fakeDescription.appendChild(brNode);
});
fakeDescription.removeChild(fakeDescription.lastChild);

fakeTitle.addEventListener("input", fakeTitleListener);
fakeTitle.addEventListener("keypress", fakeTitleListener);

fakeDescription.addEventListener("input", fakeDescriptionListener);
fakeDescription.addEventListener("keypress", fakeDescriptionListener);