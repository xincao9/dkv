wrk.method="PUT"
wrk.headers["content-type"]="application/json"

count = 1

request = function()
	local kv = string.format('{"k":"%d", "v":"%s"}', count, "新华社北京10月6日电 10月6日，中共中央总书记、国家主席习近平同朝鲜劳动党委员长、国务委员会委员长金正恩就中朝建交70周年互致贺电。  习近平在贺电中说，值此中华人民共和国和朝鲜民主主义人民共和国建交70周年之际，我谨代表中国共产党、中国政府、中国人民，向你并通过你，向朝鲜劳动党、朝鲜政府、朝鲜人民致以热烈的祝贺。  习近平表示，70年前，中朝正式建立外交关系，这在两党两国关系史上具有划时代的重要意义。朝鲜是最早同新中国建交的国家之一。70年来，在两党两国历代领导人关怀和双方共同努力下，中朝传统友谊经受住了国际风云变幻和时代变迁的考验，不断发展，历久弥坚，深入人心。两国各领域交往合作成果丰硕，不仅有力促进了两国各自社会主义事业繁荣进步，也为维护地区和平稳定发挥了重要积极作用。  习近平强调，中朝传统友谊是两党、两国、两国人民共同的宝贵财富。维护好、巩固好、发展好中朝关系，始终是中国党和政府坚定不移的方针。我高度重视中朝关系发展，珍视同委员长同志的互信和友谊。去年以来，我同委员长同志五次会晤，达成一系列重要共识，共同引领中朝关系进入新的历史时期。中方愿同朝方携手努力，以建交70周年为契机，推动中朝关系长期健康稳定发展，更好造福两国和两国人民。  金正恩在贺电中表示，值此朝中建交70周年之际，我谨代表朝鲜劳动党、朝鲜民主主义人民共和国政府和朝鲜人民，向总书记同志，并通过总书记同志向中国共产党、中华人民共和国政府和全体中国人民致以最热烈的祝贺和最诚挚的祝愿。  金正恩表示，朝中两国建交具有划时代意义。70年来，朝中两党、两国人民在维护和发展社会主义事业的征程中始终同生死、共患难，历经岁月洗礼，书写了伟大的朝中友谊历史。当前，朝中关系步入继往开来、承前启后的重要关键时期。坚决继承朝中友谊这一优秀传统，实现两国友好合作关系的全面复兴，是我和朝鲜党、政府坚定不移的立场。我愿同总书记同志紧密携手，按照朝中两国人民的共同愿望，巩固和发展令世界羡慕的朝中友谊，用友好和团结的力量坚定维护社会主义事业，坚定维护朝鲜半岛和世界的和平与稳定。")
	count = count + 1
	if count > 10000000 then
		count=0
	end
	return wrk.format(nil, nil, nil, kv)
end
