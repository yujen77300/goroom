const signoutBtn = document.getElementById("signout-btn")
const createRoomBtn = document.querySelector(".create-room-btn")
const memberName = document.getElementById("member-name")
const roomInput = document.getElementById("room-input")
const joinBtn = document.getElementById("join-btn")
const joinWrongInput = document.getElementById("join-wrong-input")
const avatarUpdateBtn = document.querySelector('.avatar-update-button')
const hiddenInput = document.querySelector('.avatar-update-input')
const avatar = document.querySelector('.avatar')
const avatarHint = document.querySelector('.avatar-hint')
let accountEmail = ""
nameOnNavbar()
getUserAvatar()

avatarUpdateBtn.addEventListener('click', function () {
  hiddenInput.click()
  hiddenInput.addEventListener('change', function (e) {
    let form = new FormData();
    console.log(e.target.files[0].type)
    let fileType = e.target.files[0].type
    if ((fileType == "image/jpeg") || (fileType == "image/png")) {
      form.append('avatarUrl', e.target.files[0])
      form.append('accountEmail', accountEmail)
      let object = {};
      form.forEach((val, key) => {
        object[key] = val;
      });
      console.log(object)
      avatarHint.style.display = "none"
      updateAvatar(form)
    }else{
      avatarHint.style.display="block"
    }

  })
})


signoutBtn.addEventListener("click", () => {
  fetch(
    "/api/user/auth"
  ).then(function (response) {
    return response.json();
  }).then(function (data) {
    if (data.data != undefined) {
      deleteAccount()
    }
  })
});

function nameOnNavbar() {
  fetch(
    "/api/user/auth"
  ).then(function (response) {
    return response.json();
  }).then(function (data) {
    memberName.textContent = "Hello, " + data.data.name
  })
}

createRoomBtn.addEventListener("click", () => {
  document.location.href = "/room/create"
})

async function deleteAccount() {
  let url = "/api/user/auth"
  let options = {
    method: "DELETE",
  }
  try {
    let response = await fetch(url, options);
    if (response.status === 200) {
      document.location.href = '/'
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

roomInput.addEventListener("input", (e) => {
  if (roomInput.value === '') {
    joinBtn.style.display = "block"
    joinBtn.style.color = "#666666";
    joinWrongInput.style.display = "none"
  } else if (/^[a-zA-Z0-9-]+$/.test(e.target.value)) {
    joinBtn.style.display = "block"
    joinBtn.style.color = "#1158bd";
    roomInput.style.border = "2px #1158bd solid";
    joinWrongInput.style.display = "none"
  }
  else {
    joinBtn.style.color = "#666666";
    roomInput.style.border = "1px red solid";
    joinWrongInput.style.display = "block"
    joinBtn.style.display = "none"
  }
})

roomInput.addEventListener("click", () => {
  joinBtn.style.display = "block"
  joinBtn.style.color = "#666666";
  joinWrongInput.style.display = "none"
})

joinBtn.addEventListener("click", () => {
  document.location.href = `/room/${roomInput.value}`
})


async function getUserAvatar() {
  let url = "/api/user/avatar"
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      avatar.style.backgroundImage = `url(${result.userAvatar})`
      accountEmail = result.user
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

async function updateAvatar(form) {
  let url = "api/user/avatar"
  let options = {
    body: form,
    method: "POST",
    header: "test"
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      avatar.style.backgroundImage = `url(${result.newAvatarUrl})`
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}