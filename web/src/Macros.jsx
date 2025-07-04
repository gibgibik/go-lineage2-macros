import {Box, Button, ButtonGroup, Container, TextField} from "@mui/material";
import {MacrosAction} from "./MacrosAction.jsx";

const items = []
const INPUT_COUNT = 10;
for (let i = 0; i < INPUT_COUNT; i++) {
    items.push(<Box sx={{display: 'flex', gap: 2, m: 2}} key={i}>
        <MacrosAction name={'action' + i}/>
        <TextField variant={"outlined"} fullWidth={true} name={'details' + i} label={'Details'}/>
        <TextField variant={"outlined"} fullWidth={true} name={'period_seconds' + i} label={'Period Seconds'}/>
    </Box>);
}
export const Macros = () => {
    return (<Container>
        {items}
        <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
            <Button>Save</Button>
        </ButtonGroup>
    </Container>)
        ;
}
