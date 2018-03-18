function postData(url, data) {
    return fetch(url, {
        body: data,
        cache: 'no-cache',
        credentials: 'same-origin',
        method: 'POST',
        mode: 'cors',
        redirect: 'follow'
    }).then(response => response.json())
}

const fakeTitle = document.querySelector(".title.fake-input[contenteditable]");
const fakeDescription = document.querySelector(".description.fake-input[contenteditable]");

const actualTitle = document.querySelector(".update-form input[name=title]");
const actualDescription = document.querySelector(".update-form input[name=description]");

const save = document.querySelector("#save");
const updateForm = document.querySelector(".update-form");

let lastTimeOut = null;
let lastSaved = null;

const currentState = () => {
    const data = new FormData(document.forms.namedItem("upload"));
    data.append("from_js", "true");
};

const doSave = () => {
    const data = currentState();
    save.value = "Saving…";
    postData(location.href, data).then((json) => {
        save.value = "Saved";
        lastSaved = data;
    })
};

const scheduleSave = () => {
    if (lastTimeOut !== null) {
        clearTimeout(lastTimeOut);
    }
    lastTimeOut = setTimeout(doSave, 300)
};

const fakeTitleListener = (event) => {
    requestAnimationFrame(() => {
        document.title = event.target.innerText + " | i.k8r";
        actualTitle.value = fakeTitle.innerText;
    });
    scheduleSave();

};
const fakeDescriptionListener = (event) => {
    requestAnimationFrame(() => {
        actualDescription.value = fakeDescription.innerText;
    });
    scheduleSave();
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

save.addEventListener("click", (e) => {
    e.preventDefault();

    doSave();
});

window.addEventListener("beforeunload", (e) => {
    const state = currentState();
    if (lastSaved !== null && lastSaved !== state) {
        const message = "Your changes have not been saved. Are you sure you want to leave?";
        e.preventDefault();
        e.returnValue = message;
        return message;
    }
};