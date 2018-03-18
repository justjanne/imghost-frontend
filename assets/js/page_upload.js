function postData(url, data) {
    return fetch(url, {
        body: data,
        cache: 'no-cache',
        credentials: 'same-origin',
        method: 'POST',
        mode: 'cors',
        redirect: 'follow'
    }).then(response => response.json());
}
const form = document.querySelector('form.upload');
const element = document.querySelector('form.upload input[type=file]');
const results = document.querySelector('.uploading.images');
element.addEventListener('change', () => {
    for (let file of element.files) {
        const reader = new FileReader();
        reader.addEventListener('load', e => {
            const dataUrl = e.target.result;
            const node = function () {
                var $$a = document.createElement('div');
                $$a.setAttribute('class', 'uploading image');
                return $$a;
            }.call(this);
            results.appendChild(node);
            const data = new FormData();
            data.append('file', file, file.name);
            postData('/upload/', data).then(json => {
                if (json.success) {
                    node.querySelector('a.image').href = '/' + json.id;
                    node.querySelector('a.image img').src = '/' + json.id;
                } else {
                    node.insertBefore(function () {
                        var $$b = document.createElement('div');
                        $$b.setAttribute('class', 'alert error');
                        $$b.appendChildren(JSON.stringify(json.errors));
                        return $$b;
                    }.call(this), node.querySelector('.description'));
                }
                console.log(json);
            });
        });
        reader.readAsDataURL(file);
    }
});