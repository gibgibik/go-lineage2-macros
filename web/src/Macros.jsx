import {Box, Button, ButtonGroup, Container, TextField} from "@mui/material";
import {MacrosAction} from "./MacrosAction.jsx";
import {getProfile, saveProfile} from "./api.js";
import {useState} from "react";

const items = []
const INPUT_COUNT = 10;
const PROFILE_NAME = 'static_profile';//@make few profiles?
for (let i = 0; i < INPUT_COUNT; i++) {
    items.push(<Box sx={{display: 'flex', gap: 2, m: 2}} key={i}>
        <MacrosAction name={'action[]'}/>
        <TextField variant={"outlined"} fullWidth={true} name={'details[]'} label={'Details'}/>
        <TextField variant={"outlined"} fullWidth={true} name={'period_seconds[]'} label={'Period Seconds'}/>
    </Box>);
}
getProfile(PROFILE_NAME);
export const Macros = () => {
    const [submitDisabled, disableSubmit] = useState(false);
    const handleSubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        // const data = Object.fromEntries(formData.entries());
        // console.log(Object.fromEntries(formData.entries()))
        const obj = {};
        for (const [key, value] of formData.entries()) {
            if (key.endsWith('[]')) {
                const cleanKey = key.slice(0, -2);
                if (!obj[cleanKey]) obj[cleanKey] = [];
                obj[cleanKey].push(value);
            } else {
                obj[key] = value;
            }
        }
        obj['profile'] = PROFILE_NAME

        disableSubmit(true);
        await saveProfile(PROFILE_NAME, obj);
        disableSubmit(false);
    }
    return (<Container>
        <form onSubmit={handleSubmit}>
            {items}
            <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                <Button type={"submit"} disabled={submitDisabled}>Save</Button>
            </ButtonGroup>
        </form>
    </Container>)
        ;
}
