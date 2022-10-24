const id = document.getElementById('id')
const password = document.getElementById('password')
const login = document.getElementById('login')
let errStack = 0;
var ws;

login.addEventListener('click', () => {

    var obj = new Object();
    obj.msgtype = "LOGIN"
    obj.userid = id.value;
    obj.userpw = password.value;

    ws = new WebSocket('ws://' + window.location.host + '/ws');
    ws.onopen = function(event) {
        ws.send(JSON.stringify(obj));
    }

    ws.onmessage = function (event) {
        console.log(event.data);
        recData = JSON.parse(event.data);
        if (recData == "LOGIN_SUCC") {
            alert('로그인 되었습니다!')
            location.replace('./main.html')
        }
        else if (recData == "LOGIN_FAIL") {
            alert('아이디와 비밀번호를 다시 한 번 확인해주세요!')
            errStack ++;
        }
        else {
            alert("몰라")
        }
    }
})