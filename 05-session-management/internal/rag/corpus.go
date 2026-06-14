package rag

type CorpusEntry struct {
	Source  string
	Content string
}

// KantoCorpus returns Season 1 Kanto knowledge snippets for embedding.
func KantoCorpus() []CorpusEntry {
	return []CorpusEntry{
		{Source: "pokedex/001", Content: "Bulbasaur is a Grass/Poison starter from Pallet Town. It carries a plant bulb on its back that grows as it absorbs sunlight."},
		{Source: "pokedex/004", Content: "Charmander is a Fire-type starter. The flame on its tail shows its life force; a steady flame means good health."},
		{Source: "pokedex/007", Content: "Squirtle is a Water-type starter. It withdraws into its shell and sprays water from its mouth in battle."},
		{Source: "pokedex/010", Content: "Caterpie is a Bug-type found in Viridian Forest. It evolves into Metapod, then Butterfree."},
		{Source: "pokedex/013", Content: "Weedle is a Bug/Poison Pokémon common in Kanto forests. It has a poisonous stinger on its head."},
		{Source: "pokedex/016", Content: "Pidgey is a Normal/Flying Pokémon seen on Route 1. It is timid and often hides in tall grass."},
		{Source: "pokedex/019", Content: "Rattata is a Normal-type found in cities and routes across Kanto. It gnaws on anything."},
		{Source: "pokedex/025", Content: "Pikachu is an Electric-type mouse Pokémon. It stores electricity in its cheek pouches and is common near forests in Kanto."},
		{Source: "pokedex/035", Content: "Clefairy is a Fairy-type associated with Mt. Moon and moonlight. It is rarely seen in the wild."},
		{Source: "pokedex/039", Content: "Jigglypuff is a Normal/Fairy Pokémon that sings a lullaby to make opponents sleep."},
		{Source: "pokedex/041", Content: "Zubat is a Poison/Flying Pokémon that lives in caves such as Mt. Moon and uses supersonic waves."},
		{Source: "pokedex/050", Content: "Diglett is a Ground-type that burrows underground, leaving only its head visible."},
		{Source: "pokedex/054", Content: "Psyduck is a Water-type that suffers headaches which trigger psychic powers."},
		{Source: "pokedex/058", Content: "Growlithe is a Fire-type loyal to its Trainer, often used by Officer Jenny in Kanto."},
		{Source: "pokedex/063", Content: "Abra is a Psychic-type that teleports away when startled. It sleeps eighteen hours a day."},
		{Source: "pokedex/066", Content: "Machop is a Fighting-type that trains by lifting heavy objects to build muscle."},
		{Source: "pokedex/074", Content: "Geodude is a Rock/Ground Pokémon found on mountain paths. It looks like a round rock."},
		{Source: "pokedex/092", Content: "Gastly is a Ghost/Poison-type formed mostly of gas. It can slip through any opening."},
		{Source: "pokedex/095", Content: "Onix is a Rock/Ground serpent made of boulders. Brock uses an Onix as his main Pokémon."},
		{Source: "pokedex/129", Content: "Magikarp is a Water-type known for weak attacks but it evolves into the powerful Gyarados."},
		{Source: "pokedex/133", Content: "Eevee is a Normal-type with unstable genetics, allowing multiple evolution paths."},
		{Source: "pokedex/143", Content: "Snorlax is a Normal-type that blocks paths while sleeping and eats nearly a ton of food daily."},
		{Source: "pokedex/147", Content: "Dratini is a rare Dragon-type found near water. It sheds its skin as it grows."},
		{Source: "pokedex/150", Content: "Mewtwo is a Psychic-type Legendary Pokémon created from Mew's DNA in a lab on Cinnabar Island."},
		{Source: "location/viridian-forest", Content: "Viridian Forest lies between Viridian City and Pewter City. Bug-types such as Caterpie, Weedle, and Pikachu appear among the trees."},
		{Source: "location/route-1", Content: "Route 1 connects Pallet Town to Viridian City. Young trainers often encounter Pidgey and Rattata here."},
		{Source: "location/pewter-city", Content: "Pewter City is home to Brock, the Rock-type Gym Leader, and the Pewter Museum of Science."},
		{Source: "location/cerulean-city", Content: "Cerulean City has Misty, the Water-type Gym Leader, and a famous bridge north of town."},
		{Source: "location/mt-moon", Content: "Mt. Moon is a cave route between Pewter and Cerulean. Trainers find Zubat, Clefairy, and fossils there."},
		{Source: "location/celadon-city", Content: "Celadon City hosts Erika's Grass Gym and the Celadon Department Store."},
		{Source: "location/saffron-city", Content: "Saffron City is dominated by Silph Co. and Sabrina's Psychic Gym."},
		{Source: "location/cinnabar-island", Content: "Cinnabar Island has a Fire Gym led by Blaine and an abandoned Pokémon Lab."},
		{Source: "type/electric", Content: "Electric-type Pokémon are strong against Water and Flying types. They are weak to Ground-type moves."},
		{Source: "type/grass", Content: "Grass-type Pokémon thrive in forests and are strong against Water, Ground, and Rock types."},
		{Source: "type/fire", Content: "Fire-type Pokémon are strong against Grass, Ice, Bug, and Steel types but weak to Water and Rock."},
		{Source: "type/water", Content: "Water-type Pokémon are common near rivers and coasts. They are strong against Fire, Ground, and Rock."},
		{Source: "anime/ash", Content: "Ash Ketchum of Pallet Town began his journey at age ten with a Pikachu given by Professor Oak."},
		{Source: "anime/oak", Content: "Professor Oak gives starter Pokémon to new trainers from his lab in Pallet Town."},
		{Source: "anime/team-rocket", Content: "Team Rocket, led by Giovanni, often tries to steal rare Pokémon such as Pikachu in the Kanto region."},
	}
}
