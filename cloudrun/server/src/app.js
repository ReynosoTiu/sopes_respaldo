const express = require('express');
const { MongoClient } = require('mongodb');
const dotenv = require('dotenv');
const cors = require('cors');

dotenv.config();

const app = express();
const port = process.env.PORT || 3000;
const uri = process.env.MONGO_URI || "mongodb://localhost:27017"; 
app.use(cors());

const client = new MongoClient(uri);

app.get('/logs', async (_req, res) => {
    try {
        await client.connect();
        const database = client.db("logs");
        const logs = database.collection("vote-logs");

        const query = logs.find({}).sort({ timestamp: -1 }).limit(20);
        const results = await query.toArray();

        res.status(200).json(results);
    } catch (e) {
        res.status(500).json({ message: e.message });
    } finally {
        await client.close();
    }
});

app.listen(port, () => {
    console.log(`Server running at http://localhost:${port}`);
});
