const setupEditor = () => {
    var editor = ace.edit("editor");
    editor.setTheme("ace/theme/monokai");
    editor.getSession().setMode("ace/mode/c_cpp");
    return editor;
}

const postCallToWAF = async (code) => {
    const rawResponse = await fetch('/waf', {
        method: 'POST',
        headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
        },
        body: JSON.stringify({code})
    });
    const content = await rawResponse.json();
    localStorage.setItem('waf-token', content.token);
    console.log(content);
    document.getElementById("output").innerHTML = content.output;
}

const setupRunClick = (editor) => {
    document.getElementById("run-btn").addEventListener("click", () => postCallToWAF(editor.getValue()))
}

const setupDownloadClick = () => {
    document.getElementById("download-btn").addEventListener("click", () => window.open(`/download?token=${localStorage.getItem('waf-token')}`, '_blank').focus())
}

const editor = setupEditor();
setupRunClick(editor);
setupDownloadClick();
