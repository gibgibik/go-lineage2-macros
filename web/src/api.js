import axios from 'axios';

const api = axios.create({
    baseURL: 'http://localhost:8088/api',
});

export const getProfile = (profileName) => {
    return api.get('/profile/' + profileName).then((response) => {
        console.log(response)
    }).catch((error) => {
        console.log(error);
    })
}

export const saveProfile = (profileName, formData) => {
    return api.post('/profile/' + profileName, formData).then((response) => {
        console.log(response)
    })
}
