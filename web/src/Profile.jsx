import React, {useState} from "react";
import {Box, Button, List, ListItem, ListItemText, Stack, TextField, Typography,} from "@mui/material";

export const Profiles = ({profilesList = [], setProfile, setProfilesList}) => {
    const [loadedFileId, setLoadedFileId] = useState(null);
    const [newFileName, setNewFileName] = useState("");

    const handleLoad = (profile) => {
        setLoadedFileId(profile);
        setProfile(profile);
    };

    const handleAddFile = () => {
        const trimmedName = newFileName.trim();
        if (!trimmedName) return;

        const newProfile = trimmedName;
        setProfilesList(() => [...profilesList, newProfile]);
        setNewFileName("");
    };

    return (
        <Box sx={{maxWidth: 500, mx: "auto", mt: 4}}>
            <Typography variant="h6" gutterBottom>
                Profiles
            </Typography>

            <Stack direction="row" spacing={2} mb={2}>
                <TextField
                    label="Create new profile"
                    variant="outlined"
                    size="small"
                    value={newFileName}
                    onChange={(e) => setNewFileName(e.target.value)}
                    fullWidth
                />
                <Button variant="contained" onClick={handleAddFile}>
                    Add
                </Button>
            </Stack>

            <List>
                {profilesList.map((profileName) => (
                    <ListItem
                        key={profileName}
                        sx={{
                            // bgcolor: profileName === loadedFileId ? "indigo" : "inherit",
                            border: profileName === loadedFileId ? 1 : 'none',
                            borderColor: 'green',
                            borderRadius: 1,
                            mb: 1,
                        }}
                        secondaryAction={
                            <Button variant="outlined" onClick={() => handleLoad(profileName)}>
                                Load
                            </Button>
                        }
                    >
                        <ListItemText primary={profileName}/>
                    </ListItem>
                ))}
            </List>
        </Box>
    );
}
