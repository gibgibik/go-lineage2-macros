import {Box, Button, ButtonGroup, Container, TextField} from "@mui/material";
import {MacrosAction} from "./MacrosAction.jsx";
import {getProfile, saveProfile} from "./api.js";

const items = []
const INPUT_COUNT = 10;
const PROFILE_NAME = 'static_profile';//@make few profiles?
for (let i = 0; i < INPUT_COUNT; i++) {
    items.push(<Box sx={{display: 'flex', gap: 2, m: 2}} key={i}>
        <MacrosAction name={'action' + i}/>
        <TextField variant={"outlined"} fullWidth={true} name={'details' + i} label={'Details'}/>
        <TextField variant={"outlined"} fullWidth={true} name={'period_seconds' + i} label={'Period Seconds'}/>
    </Box>);
}
getProfile(PROFILE_NAME);
export const Macros = () => {
    const handleSubmit = (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());
        saveProfile(PROFILE_NAME, data);
    }
    return (<Container>
        <form onSubmit={handleSubmit}>
            {items}
            <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                <Button type={"submit"}>Save</Button>
            </ButtonGroup>
        </form>
    </Container>)
        ;
}
