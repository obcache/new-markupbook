import '../wailsjs/go/main/App';

const pagesList = document.getElementById('pages');
const titleEl = document.getElementById('title');
const htmlEl = document.getElementById('html');
const saveBtn = document.getElementById('saveBtn');
const newBtn = document.getElementById('newBtn');
const renameBtn = document.getElementById('renameBtn');
const newTitleInput = document.getElementById('newTitleInput');

let currentTitle = null;

async function refreshPages(){
    const pages = await window.backend.App.ListPages();
    pagesList.innerHTML = '';
    pages.forEach(p => {
        const li = document.createElement('li');
        li.textContent = p;
        li.onclick = async () => {
            currentTitle = p;
            titleEl.textContent = p;
            const html = await window.backend.App.LoadPage(p);
            htmlEl.value = html;
        };
        pagesList.appendChild(li);
    });
    if(pages.length){
        pagesList.querySelector('li').click();
    } else {
        titleEl.textContent = 'No pages';
        htmlEl.value = '';
    }
}

saveBtn.onclick = async () => {
    const newTitle = newTitleInput.value.trim() || currentTitle;
    if(!currentTitle){
        alert('No current page selected');
        return;
    }
    try{
        const etag = await window.backend.App.GetETag();
        await window.backend.App.SavePageIfMatch(currentTitle, newTitle, htmlEl.value, etag);
    }catch(e){
        if(String(e).includes('etag')){
            if(confirm('The document changed since you opened it. Reload now?')){
                await refreshPages();
            }
            return;
        }
        throw e;
    }
    currentTitle = newTitle;
    await refreshPages();
};

newBtn.onclick = async () => {
    const t = prompt('New page title:');
    if(!t) return;
    await window.backend.App.NewPage(t);
    await refreshPages();
};

renameBtn.onclick = async () => {
    const t = prompt('Rename to:');
    if(!t || !currentTitle) return;
    await window.backend.App.RenamePage(currentTitle, t);
    currentTitle = t;
    await refreshPages();
};

// initial load
async function startup(){
    // disable UI until authenticated
    saveBtn.disabled = true;
    newBtn.disabled = true;
    renameBtn.disabled = true;

    // simple prompt-based auth loop
    while(true){
        const token = prompt('Enter MARKUPBOOK token:');
        if(token === null) {
            // user cancelled; leave UI disabled
            return;
        }
        try{
            await window.backend.App.Authenticate(token);
            break;
        }catch(e){
            alert('Authentication failed. Try again.');
        }
    }

    // enable UI
    saveBtn.disabled = false;
    newBtn.disabled = false;
    renameBtn.disabled = false;

    await refreshPages();
}

startup();
