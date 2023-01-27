const homeImg = document.querySelector(".home-img")
const singInCard = document.getElementById("signin-card")
const singInCardButton = document.getElementById("signin-card-button")
const signInEmail = document.getElementById("signin-email")
const signInPassword = document.getElementById("signin-password")
const signInFail = document.getElementById("signin-fail")



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
    signInFail.textContent = "請輸入帳號與密碼"
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
    console.log("測試回來的")
    console.log(response)
    console.log(response.status)
    let result = await response.json();
    console.log(result)
    if (response.status === 200) {
      // window.location.reload();
      console.log("成功了")
    } else if (response.status === 400) {
      // loginFail.textContent = result.message;
      signInEmail.value = ""
      signInPassword.value = ""
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}