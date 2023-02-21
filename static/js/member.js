const signoutBtn = document.getElementById("signout-btn")
const createRoomBtn = document.querySelector(".create-room-btn")
const memberName = document.getElementById("member-name")
const roomInput = document.getElementById("room-input")
const joinBtn = document.getElementById("join-btn")
const joinWrongInput = document.getElementById("join-wrong-input")
const avatarUpdateBtn = document.querySelector('.avatar-update-button')
const hiddenInput = document.querySelector('.avatar-update-input')
const avatar = document.querySelector('.avatar')
let avatarUrl = ""
nameOnNavbar()
updateUserAvatar()

avatarUpdateBtn.addEventListener('click', function () {
  console.log(hiddenInput)
  hiddenInput.click()
  // hiddenInput.addEventListener('change', function (e) {
  //   let form = new FormData();
  //   form.append('form', e.target.files[0])
  //   uploadImage(form)
  // })
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


async function updateUserAvatar() {
  let url = "/api/user/avatar"
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      avatar.style.backgroundImage = `url(${result.userAvatar})`
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}