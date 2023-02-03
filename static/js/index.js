const homeImg = document.querySelector(".home-img")
const signInCard = document.getElementById("signin-card")
const singInCardButton = document.getElementById("signin-card-button")
const signInEmail = document.getElementById("signin-email")
const signInPassword = document.getElementById("signin-password")
const inputHint = document.getElementById("input-hint")
const signInHint = document.getElementById("signin-hint")
const signinBtn = document.querySelectorAll(".signin-btn")
const signUpBtn = document.getElementById("signup-btn")
const signUpCard = document.getElementById("signup-card")
const signUpCardButton = document.getElementById("signup-card-button")
const signUpName = document.getElementById("signup-username")
const signUpEmail = document.getElementById("signup-email")
const signUpPassword = document.getElementById("signup-password")


signinBtn.forEach((button) => {
  button.addEventListener("click", () => {
    homeImg.style.display = "none"
    signUpCard.style.display = "none"
    signInCard.style.display = "block"
  })
})

signUpBtn.addEventListener("click", () => {
  homeImg.style.display = "none"
  signInCard.style.display = "none"
  signUpCard.style.display = "block"
})

signUpCardButton.addEventListener("click", () => {
  let signUpInputName = signUpName.value
  let signUpInputEmail = signUpEmail.value
  let signUpInputPassword = signUpPassword.value
  if (signUpInputName.length != 0 && signUpInputEmail != 0 && signUpInputPassword != 0) {
    if (!emailValidation(signUpInputEmail) && !passwordValidation(signUpInputPassword)) {
      inputHint.textContent = "Invalid email and password"
      signUpEmail.value = ""
      signUpPassword.value = ""
    } else if (!emailValidation(signUpInputEmail)) {
      inputHint.textContent = "Invalid email"
      signUpEmail.value = ""
    } else if (!passwordValidation(signUpInputPassword)) {
      inputHint.textContent = "Invalid password, at least 8 characters, one number and one English letter are required"
      signUpPassword.value = ""
    } else {
      let signUpData = {
        "username": signUpInputName,
        "email": signUpInputEmail,
        "password": signUpInputPassword
      };
      signUpAccount(signUpData)

    }
  } else {
    inputHint.textContent = "Username, email or password is empty"
  }
})


singInCardButton.addEventListener("click", () => {
  const signInInputEmail = signInEmail.value
  const signInInputPassword = signInPassword.value
  if (signInInputEmail.length != 0 && signInInputPassword != 0) {
    const signInData = {
      "email": signInInputEmail,
      "password": signInInputPassword
    };
    console.log(signInData)
    signInAccount(signInData)
  } else {
    console.log("近來錯誤")
    signInHint.textContent = "Email or password is empty"
  }
})

async function signInAccount(data) {
  let url = "/api/user/auth"
  let options = {
    method: "PUT",
    body: JSON.stringify(data),
    headers: {
      "Content-type": "application/json",
    }
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      document.location.href = '/member'

    } else if (response.status === 401) {
      signInHint.textContent = result.message;
      signInEmail.value = ""
      signInPassword.value = ""
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

async function signUpAccount(data) {
  let url = "/api/user"
  let options = {
    method: "POST",
    body: JSON.stringify(data),
    headers: {
      "Content-type": "application/json",
    }
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      inputHint.textContent = "Success! Please Sign in"
    } else if (response.status === 400) {
      inputHint.textContent = result.message;
      signUpName.value = ""
      signUpEmail.value = ""
      signUpPassword.value = ""
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

// Regular Expression驗證信箱
function emailValidation(email) {
  if (email.search(/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/) != -1) {
    return true
  } else {
    return false
  }
}
// Regular Expression驗證密碼至少包含數字、英文字母
function passwordValidation(password) {
  if (password.search(/^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{8,}$/) != -1) {
    return true
  } else {
    return false
  }
}